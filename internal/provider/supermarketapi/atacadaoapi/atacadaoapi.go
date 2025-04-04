package atacadaoapi

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/danielmesquitta/supermarket-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/errs"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/supermarketapi"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

type AtacadaoAPI struct {
	c *resty.Client
}

func New(
	e *env.Env,
) *AtacadaoAPI {
	c := resty.New().SetBaseURL(e.AtacadaoAPIBaseURL)

	return &AtacadaoAPI{
		c: c,
	}
}

var categories = []string{
	"bebidas",
	"mercearia",
	"limpeza",
	"higiene-e-perfumaria",
	"padaria-e-matinais",
	"papelaria",
	"pet-shop",
	"automotivo",
	"eletronicos-e-eletroportateis",
	"vestuario",
	"utilidades-domesticas",
	"descartaveis-e-embalagens",
	"esporte-e-lazer",
}

func (a *AtacadaoAPI) ListProducts(
	ctx context.Context,
) ([]entity.Product, error) {
	products := []entity.Product{}

	mu := &sync.Mutex{}
	g := errgroup.Group{}
	g.SetLimit(5)
	for _, category := range categories {
		g.Go(func() error {
			return a.bulkRequests(ctx, mu, &products, category)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, errs.New(err)
	}

	return products, nil
}

func (a *AtacadaoAPI) bulkRequests(
	ctx context.Context,
	mu *sync.Mutex,
	products *[]entity.Product,
	category string,
) error {
	defaultOpts := requestOptions{
		Page:     1,
		Size:     100,
		Total:    0,
		Category: category,
	}
	err := a.doRequest(ctx, mu, products, defaultOpts)
	if err != nil {
		return err
	}

	totalPages := math.Ceil(
		float64(defaultOpts.Total) / float64(defaultOpts.Size),
	)

	defaultOpts.Total = int(totalPages)

	g := errgroup.Group{}
	g.SetLimit(10)
	for i := 2; i <= int(totalPages); i++ {
		g.Go(func() error {
			opts := defaultOpts
			opts.Page = i
			return a.doRequest(ctx, mu, products, opts)
		})
	}

	if err := g.Wait(); err != nil {
		return errs.New(err)
	}

	return nil
}

type requestOptions struct {
	Page     int
	Size     int
	Total    int
	Category string
}

func (a *AtacadaoAPI) doRequest(
	ctx context.Context,
	mu *sync.Mutex,
	products *[]entity.Product,
	opts requestOptions,
) error {
	queryParams, err := buildQueryParams(opts.Page, opts.Size, opts.Category)
	if err != nil {
		return errs.New(err)
	}
	res, err := a.c.R().
		SetContext(ctx).
		SetQueryParams(queryParams).
		Get("/")
	if err != nil {
		return errs.New(err)
	}
	if res.IsError() {
		return errs.New(
			fmt.Sprintf("error response: %s", res.String()),
		)
	}
	response, err := parseResponse(res)
	if err != nil {
		return errs.New(err)
	}

	mu.Lock()
	*products = append(*products, response.Products...)
	mu.Unlock()

	return nil
}

func buildQueryParams(
	page, size int,
	category string,
) (map[string]string, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	type SelectedFacet struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	type Variables struct {
		First          int             `json:"first"`
		After          string          `json:"after"`
		Sort           string          `json:"sort"`
		Term           string          `json:"term"`
		SelectedFacets []SelectedFacet `json:"selectedFacets"`
	}

	variables := Variables{
		First: size,
		After: strconv.Itoa((page - 1) * size),
		Sort:  "score_desc",
		Term:  "",
		SelectedFacets: []SelectedFacet{
			{
				Key:   "category-1",
				Value: category,
			},
			{
				Key:   "region-id",
				Value: "U1cjYXRhY2FkYW9icjMw",
			},
			{
				Key:   "channel",
				Value: "{\"salesChannel\":\"1\",\"seller\":\"atacadaobr30\",\"regionId\":\"U1cjYXRhY2FkYW9icjMw\"}",
			},
			{
				Key:   "locale",
				Value: "pt-BR",
			},
		},
	}

	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, errs.New(err)
	}

	return map[string]string{
		"operationName": "ProductsQuery",
		"variables":     string(variablesJSON),
	}, nil
}

type response struct {
	Products   []entity.Product `json:"products"`
	TotalCount int64            `json:"total_count"`
}

func parseResponse(res *resty.Response) (*response, error) {
	type Offers struct {
		HighPrice float64 `json:"highPrice"`
		LowPrice  float64 `json:"lowPrice"`
	}

	type Node struct {
		ID     string `json:"id"`
		Sku    string `json:"sku"`
		Name   string `json:"name"`
		Gtin   string `json:"gtin"`
		Offers Offers `json:"offers"`
	}

	type Edge struct {
		Node Node `json:"node"`
	}

	type PageInfo struct {
		TotalCount int64 `json:"totalCount"`
	}

	type Products struct {
		PageInfo PageInfo `json:"pageInfo"`
		Edges    []Edge   `json:"edges"`
	}

	type Search struct {
		Products Products `json:"products"`
	}

	type Data struct {
		Search Search `json:"search"`
	}

	type APIResponse struct {
		Data Data `json:"data"`
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(res.Bytes(), &apiResponse); err != nil {
		return nil, errs.New(err)
	}

	products := make(
		[]entity.Product,
		len(apiResponse.Data.Search.Products.Edges),
	)
	for i, edge := range apiResponse.Data.Search.Products.Edges {
		code := cmp.Or(edge.Node.Sku, edge.Node.Gtin)
		products[i] = entity.Product{
			Code:  &code,
			Name:  edge.Node.Name,
			Price: max(edge.Node.Offers.HighPrice, edge.Node.Offers.LowPrice),
		}
	}

	parsedResponse := &response{
		Products:   products,
		TotalCount: apiResponse.Data.Search.Products.PageInfo.TotalCount,
	}

	return parsedResponse, nil
}

var _ supermarketapi.SupermarketAPI = (*AtacadaoAPI)(nil)
