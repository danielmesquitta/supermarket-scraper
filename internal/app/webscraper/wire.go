//go:build wireinject
// +build wireinject

package webscraper

import (
	"github.com/google/wire"

	"github.com/danielmesquitta/supermarket-scraper/internal/app/webscraper/handler"
	"github.com/danielmesquitta/supermarket-scraper/internal/config"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/usecase"
	"github.com/danielmesquitta/supermarket-scraper/internal/pkg/validator"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/db/sqlite"
)

func New() *WebScraper {
	wire.Build(
		wire.Bind(new(validator.Validator), new(*validator.Validation)),
		validator.New,

		config.LoadConfig,

		usecase.NewSaveProductsUseCase,
		usecase.NewSaveErrorUseCase,

		sqlite.New,

		handler.New,

		Build,
	)
	return &WebScraper{}
}
