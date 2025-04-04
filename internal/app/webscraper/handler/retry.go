package handler

import (
	"context"
	"encoding/json"
	"log"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/errs"
)

const retryPagesLimit = 10

func (h *Handler) Retry(ctx context.Context) error {
	dbErrors, err := h.db.ListErrorsByType(
		ctx,
		errs.ErrTypeFailedProcessingProductsPage,
	)
	if err != nil {
		return errs.New(err)
	}

	if len(dbErrors) == 0 {
		return nil
	}

	defer func() {
		ids := []string{}
		for _, dbError := range dbErrors {
			ids = append(ids, dbError.ID)
		}
		if err := h.db.DeleteErrors(ctx, ids); err != nil {
			log.Fatalf("failed to delete errors: %v", err)
		}
	}()

	browser, stop, err := h.setupBrowserContext()
	if err != nil {
		return errs.New(err)
	}
	defer func() { _ = stop() }()

	g := errgroup.Group{}
	g.SetLimit(retryPagesLimit)
	for _, dbError := range dbErrors {
		g.Go(func() error {
			metadata := map[string]string{}
			if err := json.Unmarshal([]byte(dbError.Metadata), &metadata); err != nil {
				return errs.New(err)
			}

			url := metadata["page_url"]
			if url == "" {
				return errs.New("invalid metadata")
			}

			products, err := h.processProductsFromBrowserContext(
				ctx,
				browser,
				url,
			)
			if err != nil {
				return errs.New(err)
			}

			if len(products) == 0 {
				return nil
			}

			if err = h.saveProducts(ctx, products); err != nil {
				return errs.New(err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errs.New(err)
	}

	return nil
}
