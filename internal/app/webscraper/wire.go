//go:build wireinject
// +build wireinject

package webscraper

import (
	"github.com/google/wire"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/config"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/pkg/validator"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/provider/db"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/provider/db/sqlx"
)

func New() *WebScraper {
	wire.Build(
		wire.Bind(new(validator.Validator), new(*validator.Validation)),
		validator.New,

		config.LoadConfig,

		sqlx.New,
		db.New,

		Build,
	)
	return &WebScraper{}
}
