package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/app/webscraper"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/domain/errs"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	g := errgroup.Group{}
	g.Go(func() error {
		defer cancel()

		app := webscraper.New()
		if err := app.Run(ctx); err != nil {
			return err
		}

		return nil
	})

	<-ctx.Done()

	log.Println("Shutting down...")

	if err := g.Wait(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	var appErr *errs.Err
	if errors.As(err, &appErr) {
		log.Printf(
			"failed to run app: %v\nstacktrace: %v",
			appErr.Error(),
			appErr.StackTrace,
		)
		return
	}

	log.Printf("failed to run app: %v", err)
}
