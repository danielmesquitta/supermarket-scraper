package wire

import (
	"github.com/danielmesquitta/supermarket-web-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/pkg/validator"
)

func init() {
	_ = providers
	_ = devProviders
	_ = testProviders
	_ = stagingProviders
	_ = prodProviders
	_ = params
}

func params(
	v *validator.Validator,
	e *env.Env,
) {
}

var providers = []any{}

var devProviders = []any{}

var testProviders = []any{}

var stagingProviders = []any{}

var prodProviders = []any{}
