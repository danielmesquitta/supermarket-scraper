package webscraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/errs"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/provider/db"
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
	"jardinagem",
	"descartaveis-e-embalagens",
	"esporte-e-lazer",
}

type WebScraper struct {
	e  *env.Env
	db *db.DB
}

func Build(
	e *env.Env,
	db *db.DB,
) *WebScraper {
	return &WebScraper{
		e:  e,
		db: db,
	}
}

func (w *WebScraper) Run(ctx context.Context) error {
	pw, err := playwright.Run()
	if err != nil {
		return errs.New(err)
	}
	defer func() { _ = pw.Stop() }()

	cdpURL := fmt.Sprintf("http://localhost:%s", w.e.CDPPort)
	browser, err := pw.Chromium.ConnectOverCDP(cdpURL)
	if err != nil {
		return errs.New(err)
	}
	contexts := browser.Contexts()
	if len(contexts) == 0 {
		return errs.New(err)
	}
	page, err := contexts[0].NewPage()
	if err != nil {
		return errs.New(err)
	}
	defer func() { _ = page.Close() }()

	for _, category := range categories {
		err := w.processCategory(ctx, page, category)
		if err != nil {
			return errs.New(err)
		}
	}

	return nil
}

func (w *WebScraper) processCategory(
	ctx context.Context,
	page playwright.Page,
	category string,
) (err error) {
	defer func() {
		if err == nil {
			return
		}
		_ = w.saveError(
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
	products, err = w.processPage(ctx, page)
	if err != nil {
		return errs.New(err)
	}
	if len(products) == 0 {
		return nil
	}

	if err = w.saveProducts(ctx, products); err != nil {
		return errs.New(err)
	}

	pageSize := len(products)
	pagesCount := math.Ceil(float64(totalProductsCount) / float64(pageSize))

	for pageCount := 2; pageCount <= int(pagesCount); pageCount++ {
		if _, err = page.Goto(fmt.Sprintf("https://www.atacadao.com.br/%s?page=%d", category, pageCount)); err != nil {
			return errs.New(err)
		}
		if err = page.WaitForLoadState(); err != nil {
			return errs.New(err)
		}

		products, err = w.processPage(ctx, page)
		if err != nil {
			return errs.New(err)
		}
		if len(products) == 0 {
			break
		}
		if err = w.saveProducts(ctx, products); err != nil {
			return errs.New(err)
		}
	}

	return nil
}

func (w *WebScraper) processPage(
	ctx context.Context,
	page playwright.Page,
) (products []entity.Product, err error) {
	defer func() {
		if err == nil {
			return
		}
		_ = w.saveError(
			errs.New(err, errs.ErrTypeFailedProcessingProductsPage),
			map[string]any{"page_url": page.URL()},
		)
	}()

	if err = ctx.Err(); err != nil {
		return nil, errs.New(err)
	}
	productsSelector := "div[data-fs-product-listing-results='true'] section[data-testid='store-product-card-content']"
	productsLocator := page.Locator(productsSelector)

	err = productsLocator.First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		return nil, errs.New(err)
	}
	var allProductsLocator []playwright.Locator
	allProductsLocator, err = productsLocator.All()
	if err != nil {
		return nil, errs.New(err)
	}

	if len(allProductsLocator) == 0 {
		return nil, errs.New("no products found")
	}

	for _, productLocator := range allProductsLocator {
		productNameSelector := "h3"
		productNameLocator := productLocator.Locator(productNameSelector)

		var productName string
		productName, err = productNameLocator.InnerText()
		if err != nil {
			return nil, errs.New(err)
		}

		productBulkPriceSelector := ".text-lg.text-neutral-500.font-bold"
		productBulkPriceLocator := productLocator.Locator(
			productBulkPriceSelector,
		)

		var productBulkPriceStr string
		productBulkPriceStr, err = productBulkPriceLocator.InnerText()
		if err != nil {
			return nil, errs.New(err)
		}

		var productBulkPrice float64
		productBulkPrice, err = parsePrice(productBulkPriceStr)
		if err != nil {
			return nil, errs.New(err)
		}

		productPriceSelector := ".flex.items-center.gap-1"
		productPriceLocator := productLocator.Locator(productPriceSelector)

		var productPriceLocatorCount int
		productPriceLocatorCount, err = productPriceLocator.Count()
		if err != nil {
			return nil, errs.New(err)
		}

		var productPrice float64
		if productPriceLocatorCount > 0 {
			var productPriceStr string
			productPriceStr, err = productPriceLocator.InnerText()
			if err != nil {
				return nil, errs.New(err)
			}
			productPrice, err = parsePrice(productPriceStr)
			if err != nil {
				return nil, errs.New(err)
			}
		}

		actualPrice := max(productBulkPrice, productPrice)

		products = append(products, entity.Product{
			Name:  productName,
			Price: actualPrice,
		})
	}

	return products, nil
}

func (w *WebScraper) saveProducts(
	ctx context.Context,
	products []entity.Product,
) error {
	if len(products) == 0 {
		return nil
	}

	if err := w.db.CreateProducts(ctx, products); err != nil {
		return errs.New(err)
	}

	return nil
}

func (w *WebScraper) saveError(
	err error,
	metadata map[string]any,
) error {
	if err == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entityErr := entity.Error{
		Message: err.Error(),
		Type:    string(errs.ErrTypeUnknown),
	}

	var appErr *errs.Err
	if errors.As(err, &appErr) {
		if metadata != nil {
			metadataBytes, err := json.Marshal(metadata)
			if err != nil {
				return errs.New(err)
			}
			entityErr.Metadata = string(metadataBytes)
		}

		if appErr.Type != "" {
			entityErr.Type = string(appErr.Type)
		}

		entityErr.StackTrace = appErr.StackTrace
	}

	if err := w.db.CreateError(ctx, entityErr); err != nil {
		return errs.New(err)
	}

	return nil
}

func parseInt(s string) (int, error) {
	re := regexp.MustCompile("[^0-9]+")
	numStr := re.ReplaceAllString(s, "")
	return strconv.Atoi(numStr)
}

func parsePrice(s string) (float64, error) {
	centsInt, err := parseInt(s)
	if err != nil {
		return 0, errs.New(err)
	}
	price := float64(centsInt) / 100
	return price, nil
}
