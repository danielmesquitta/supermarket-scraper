package config

import (
	"github.com/danielmesquitta/supermarket-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-scraper/internal/config/log"
	"github.com/danielmesquitta/supermarket-scraper/internal/config/time"
	"github.com/danielmesquitta/supermarket-scraper/internal/pkg/validator"
)

func LoadConfig(v validator.Validator) *env.Env {
	e := env.New(v)
	log.SetDefaultLogger(e)
	time.SetServerTimeZone()

	return e
}
