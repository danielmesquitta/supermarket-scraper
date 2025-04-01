package webscraper

import (
	"context"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/app/webscraper/handler"
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
	"descartaveis-e-embalagens",
	"esporte-e-lazer",
}

type WebScraper struct {
	h *handler.Handler
}

func Build(
	h *handler.Handler,
) *WebScraper {
	return &WebScraper{
		h: h,
	}
}

func (w *WebScraper) Run(ctx context.Context) error {
	return w.h.Run(ctx)
}

func (w *WebScraper) Retry(ctx context.Context) error {
	return w.h.Retry(ctx)
}
