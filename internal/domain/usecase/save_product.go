package usecase

import (
	"context"

	"github.com/danielmesquitta/supermarket-scraper/internal/domain/entity"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/errs"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/db/sqlite"
)

type SaveProductsUseCase struct {
	db *sqlite.DB
}

func NewSaveProductsUseCase(db *sqlite.DB) *SaveProductsUseCase {
	return &SaveProductsUseCase{
		db: db,
	}
}

func (u *SaveProductsUseCase) Execute(
	ctx context.Context,
	products []entity.Product,
) error {
	if len(products) == 0 {
		return nil
	}

	names := []string{}
	for _, product := range products {
		names = append(names, product.Name)
	}

	existingProducts, err := u.db.ListProductsByNames(ctx, names)
	if err != nil {
		return errs.New(err)
	}

	existingProductsByName := map[string]entity.Product{}
	for _, product := range existingProducts {
		existingProductsByName[product.Name] = product
	}

	productsToCreate := []entity.Product{}
	for _, product := range products {
		if _, ok := existingProductsByName[product.Name]; ok {
			continue
		}
		productsToCreate = append(productsToCreate, product)
	}

	if err := u.db.CreateProducts(ctx, productsToCreate); err != nil {
		return errs.New(err)
	}

	return nil
}
