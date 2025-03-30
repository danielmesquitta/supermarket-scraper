package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/danielmesquitta/supermarket-web-scraper/internal/app/webscraper"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/config"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/pkg/errs"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/pkg/validator"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		v := validator.New()
		env := config.LoadConfig(v)

		app := webscraper.New(env)
		if err := app.Run(ctx); err != nil {
			handleError(err)
			return
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down...")

	wg.Wait()
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
