package supermarketapi

import (
	"context"

	"github.com/danielmesquitta/supermarket-scraper/internal/domain/entity"
)

type SupermarketAPI interface {
	ListProducts(ctx context.Context) ([]entity.Product, error)
}
