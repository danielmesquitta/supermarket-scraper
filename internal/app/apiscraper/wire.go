//go:build wireinject
// +build wireinject

package apiscraper

import (
	"github.com/google/wire"

	"github.com/danielmesquitta/supermarket-scraper/internal/app/apiscraper/handler"
	"github.com/danielmesquitta/supermarket-scraper/internal/config"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/usecase"
	"github.com/danielmesquitta/supermarket-scraper/internal/pkg/validator"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/db/sqlite"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/supermarketapi"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/supermarketapi/atacadaoapi"
)

func New() *APIScraper {
	wire.Build(
		wire.Bind(new(validator.Validator), new(*validator.Validation)),
		validator.New,

		config.LoadConfig,

		usecase.NewSaveProductsUseCase,
		usecase.NewSaveErrorUseCase,

		sqlite.New,

		wire.Bind(
			new(supermarketapi.SupermarketAPI),
			new(*atacadaoapi.AtacadaoAPI),
		),
		atacadaoapi.New,

		handler.New,

		Build,
	)
	return &APIScraper{}
}
