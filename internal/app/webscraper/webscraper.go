package webscraper

import (
	"github.com/danielmesquitta/supermarket-web-scraper/internal/app/webscraper/handler"
)

type WebScraper struct {
	*handler.Handler
}

func Build(
	h *handler.Handler,
) *WebScraper {
	return &WebScraper{
		Handler: h,
	}
}
