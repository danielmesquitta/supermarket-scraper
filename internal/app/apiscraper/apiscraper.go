package apiscraper

import "github.com/danielmesquitta/supermarket-web-scraper/internal/app/apiscraper/handler"

type APIScraper struct {
	h *handler.Handler
}

func Build(
	h *handler.Handler,
) *APIScraper {
	return &APIScraper{
		h: h,
	}
}
