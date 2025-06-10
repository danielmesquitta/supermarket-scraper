package handler

import (
	"context"

	"github.com/danielmesquitta/supermarket-scraper/internal/domain/errs"
)

func (h *Handler) Run(ctx context.Context) error {
	products, err := h.sa.ListProducts(ctx)
	if err != nil {
		_ = h.seuc.Execute(
			ctx,
			errs.New(err, errs.ErrTypeFailedListingProducts),
			nil,
		)
		return errs.New(err)
	}

	if err = h.spuc.Execute(ctx, products); err != nil {
		_ = h.seuc.Execute(
			ctx,
			errs.New(err, errs.ErrTypeFailedSavingProducts),
			nil,
		)
		return errs.New(err)
	}

	return nil
}
