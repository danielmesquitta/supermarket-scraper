package handler

import (
	"context"
	"fmt"
	"math"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/errs"
	"github.com/playwright-community/playwright-go"
)

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

const categoryPagesLimit = 2
const productPagesLimit = 5

func (h *Handler) Run(ctx context.Context) error {
	browser, stop, err := h.setupBrowserContext()
	if err != nil {
		return errs.New(err)
	}
	defer func() { _ = stop() }()

	g := errgroup.Group{}
	g.SetLimit(categoryPagesLimit)
	for _, category := range categories {
		g.Go(func() error {
			if err := h.processCategory(ctx, browser, category); err != nil {
				return errs.New(err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errs.New(err)
	}

	return nil
}

func (h *Handler) processCategory(
	ctx context.Context,
	browser playwright.BrowserContext,
	category string,
) (err error) {
	var page playwright.Page
	page, err = browser.NewPage()
	if err != nil {
		return errs.New(err)
	}

	defer func() {
		_ = page.Close()
		if err == nil {
			return
		}

		_ = h.saveError(
			errs.New(err, errs.ErrTypeFailedProcessingCategoryPage),
			map[string]any{"page_url": page.URL()},
		)
	}()

	if err = ctx.Err(); err != nil {
		return errs.New(err)
	}

	if _, err = page.Goto(fmt.Sprintf("https://www.atacadao.com.br/%s", category)); err != nil {
		return errs.New(err)
	}
	if err = page.WaitForLoadState(); err != nil {
		return errs.New(err)
	}

	var totalProductsCount int
	for totalProductsCount == 0 {
		time.Sleep(time.Second)

		totalProductsCountSelector := "h2[data-testid='total-product-count']"
		totalProductsCountLocator := page.Locator(totalProductsCountSelector).
			First()

		var totalProductsCountStr string
		totalProductsCountStr, err = totalProductsCountLocator.InnerText()
		if err != nil {
			return errs.New(err)
		}

		totalProductsCount, err = parseInt(totalProductsCountStr)
		if err != nil {
			return errs.New(err)
		}
	}

	var products []entity.Product
	products, err = h.processProductsFromPage(ctx, page)
	if err != nil {
		return errs.New(err)
	}
	if err = page.Close(); err != nil {
		return errs.New(err)
	}
	if len(products) == 0 {
		return nil
	}

	if err = h.saveProducts(ctx, products); err != nil {
		return errs.New(err)
	}

	pageSize := len(products)
	pagesCount := math.Ceil(float64(totalProductsCount) / float64(pageSize))

	g := errgroup.Group{}
	g.SetLimit(productPagesLimit)
	for pageCount := 2; pageCount <= int(pagesCount); pageCount++ {
		g.Go(func() error {
			url := fmt.Sprintf(
				"https://www.atacadao.com.br/%s?page=%d",
				category,
				pageCount,
			)

			products, err = h.processProductsFromBrowserContext(
				ctx,
				browser,
				url,
			)
			if err != nil {
				return errs.New(err)
			}

			if len(products) == 0 {
				return nil
			}

			if err = h.saveProducts(ctx, products); err != nil {
				return errs.New(err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errs.New(err)
	}

	return nil
}
