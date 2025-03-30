package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()

	<-ctx.Done()
	stop()

	log.Println("Shutting down...")

	wg.Wait()
}
