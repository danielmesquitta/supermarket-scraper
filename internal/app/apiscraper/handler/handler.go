package handler

import (
	"github.com/danielmesquitta/supermarket-scraper/internal/config/env"
	"github.com/danielmesquitta/supermarket-scraper/internal/domain/usecase"
	"github.com/danielmesquitta/supermarket-scraper/internal/provider/supermarketapi"
)

type Handler struct {
	e    *env.Env
	sa   supermarketapi.SupermarketAPI
	spuc *usecase.SaveProductsUseCase
	seuc *usecase.SaveErrorUseCase
}

func New(
	e *env.Env,
	sa supermarketapi.SupermarketAPI,
	spuc *usecase.SaveProductsUseCase,
	seuc *usecase.SaveErrorUseCase,
) *Handler {
	return &Handler{
		e:    e,
		sa:   sa,
		spuc: spuc,
		seuc: seuc,
	}
}
