package handler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/danielmesquitta/supermarket-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/errs"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/usecase"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/db/sqlite"
	"github.com/playwright-community/playwright-go"
)

type Handler struct {
	e    *env.Env
	db   *sqlite.DB
	spuc *usecase.SaveProductsUseCase
	seuc *usecase.SaveErrorUseCase
}

func New(
	e *env.Env,
	db *sqlite.DB,
	spuc *usecase.SaveProductsUseCase,
	seuc *usecase.SaveErrorUseCase,
) *Handler {
	return &Handler{
		e:    e,
		db:   db,
		spuc: spuc,
		seuc: seuc,
	}
}

func (h *Handler) setupBrowserContext() (browserContext playwright.BrowserContext, stop func() error, err error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, errs.New(err)
	}

	cdpURL := fmt.Sprintf("http://localhost:%s", h.e.CDPPort)
	browser, err := pw.Chromium.ConnectOverCDP(cdpURL)
	if err != nil {
		return nil, nil, errs.New(err)
	}
	contexts := browser.Contexts()
	if len(contexts) == 0 {
		return nil, nil, errs.New(err)
	}
	browserContext = contexts[0]
	return browserContext, pw.Stop, nil
}

func (h *Handler) processProductsFromBrowserContext(
	ctx context.Context,
	browser playwright.BrowserContext,
	url string,
) (products []entity.Product, err error) {
	if err = ctx.Err(); err != nil {
		return nil, errs.New(err)
	}

	var page playwright.Page
	page, err = browser.NewPage()
	if err != nil {
		return nil, errs.New(err)
	}

	defer func() {
		_ = page.Close()
		if err == nil {
			return
		}
		_ = h.seuc.Execute(
			ctx,
			errs.New(err, errs.ErrTypeFailedProcessingProductsPage),
			map[string]any{"page_url": page.URL()},
		)
	}()

	if _, err = page.Goto(url); err != nil {
		return nil, errs.New(err)
	}
	if err = page.WaitForLoadState(); err != nil {
		return nil, errs.New(err)
	}

	products, err = h.processProductsFromPage(ctx, page)
	if err != nil {
		return nil, errs.New(err)
	}

	return products, nil
}

func (h *Handler) processProductsFromPage(
	ctx context.Context,
	page playwright.Page,
) (products []entity.Product, err error) {
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
