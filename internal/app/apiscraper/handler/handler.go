package handler

import (
	"github.com/danielmesquitta/supermarket-scraper/internal/config/env"
)

type Handler struct {
	e *env.Env
}

func New(
	e *env.Env,
) *Handler {
	return &Handler{
		e: e,
	}
}
