package webscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/pkg/errs"
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
	e *env.Env
}

func New(
	e *env.Env,
) *WebScraper {
	return &WebScraper{
		e: e,
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
		err := processCategory(ctx, page, category)
		if err != nil {
			return errs.New(err)
		}
	}

	return nil
}

func processCategory(
	ctx context.Context,
	page playwright.Page,
	category string,
) error {
	if err := ctx.Err(); err != nil {
		return errs.New(err)
	}

	if _, err := page.Goto(fmt.Sprintf("https://www.atacadao.com.br/%s", category)); err != nil {
		return errs.New(err)
	}
	if err := page.WaitForLoadState(); err != nil {
		return errs.New(err)
	}

	var totalProductsCount int
	for totalProductsCount == 0 {
		time.Sleep(time.Second)

		totalProductsCountSelector := "h2[data-testid='total-product-count']"
		totalProductsCountLocator := page.Locator(totalProductsCountSelector).
			First()
		totalProductsCountStr, err := totalProductsCountLocator.InnerText()
		if err != nil {
			return errs.New(err)
		}

		totalProductsCount, err = parseInt(totalProductsCountStr)
		if err != nil {
			return errs.New(err)
		}
	}

	products, err := processPage(ctx, page, category)
	if err != nil {
		return errs.New(err)
	}
	if len(products) == 0 {
		return nil
	}

	if err := saveProductsToJSONFile(products); err != nil {
		return errs.New(err)
	}

	pageSize := len(products)
	pagesCount := math.Ceil(float64(totalProductsCount) / float64(pageSize))

	for pageCount := 2; pageCount <= int(pagesCount); pageCount++ {
		if _, err = page.Goto(fmt.Sprintf("https://www.atacadao.com.br/%s?page=%d", category, pageCount)); err != nil {
			return errs.New(err)
		}
		if err := page.WaitForLoadState(); err != nil {
			return errs.New(err)
		}

		products, err := processPage(ctx, page, category)
		if err != nil {
			return errs.New(err)
		}
		if len(products) == 0 {
			break
		}
		if err := saveProductsToJSONFile(products); err != nil {
			return errs.New(err)
		}
	}

	return nil
}

func processPage(
	ctx context.Context,
	page playwright.Page,
	category string,
) ([]entity.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, errs.New(err)
	}
	productsSelector := "div[data-fs-product-listing-results='true'] section[data-testid='store-product-card-content']"
	productsLocator := page.Locator(productsSelector)

	err := productsLocator.First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		return nil, errs.New(err)
	}

	allProductsLocator, err := productsLocator.All()
	if err != nil {
		return nil, errs.New(err)
	}

	if len(allProductsLocator) == 0 {
		return nil, errs.New("no products found")
	}

	products := []entity.Product{}
	for _, productLocator := range allProductsLocator {
		productNameSelector := "h3"
		productNameLocator := productLocator.Locator(productNameSelector)
		productName, err := productNameLocator.InnerText()
		if err != nil {
			return nil, errs.New(err)
		}

		productBulkPriceSelector := ".text-lg.text-neutral-500.font-bold"
		productBulkPriceLocator := productLocator.Locator(
			productBulkPriceSelector,
		)
		productBulkPriceStr, err := productBulkPriceLocator.InnerText()
		if err != nil {
			return nil, errs.New(err)
		}

		productBulkPrice, err := parsePrice(productBulkPriceStr)
		if err != nil {
			return nil, errs.New(err)
		}

		productPriceSelector := ".flex.items-center.gap-1"
		productPriceLocator := productLocator.Locator(productPriceSelector)
		productPriceLocatorCount, err := productPriceLocator.Count()
		if err != nil {
			return nil, errs.New(err)
		}

		var productPrice float64
		if productPriceLocatorCount > 0 {
			productPriceStr, err := productPriceLocator.InnerText()
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
			Name:     productName,
			Price:    actualPrice,
			Category: category,
		})
	}

	return products, nil
}

func saveProductsToJSONFile(
	products []entity.Product,
) error {
	filePath := path.Join("data", "products.json")

	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return errs.New(err)
		}
		if err := os.MkdirAll(path.Dir(filePath), os.ModePerm); err != nil {
			return errs.New(err)
		}
		if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
			return errs.New(err)
		}
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return errs.New(err)
	}

	var existingProducts []entity.Product
	if err := json.Unmarshal(fileData, &existingProducts); err != nil {
		return errs.New(err)
	}

	products = append(existingProducts, products...)
	jsonData, err := json.MarshalIndent(products, "", "  ")
	if err != nil {
		return errs.New(err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
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
