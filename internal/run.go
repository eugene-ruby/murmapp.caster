package internal

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"murmapp.caster/internal/app"
	"murmapp.caster/internal/config"
	"murmapp.caster/internal/server"
)

func Run() error {
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.New(*conf)
	if err != nil {
		return err
	}
	defer a.Shutdown()

	go func() {
		if err := server.StartHealthzServer(ctx, conf.AppPort); err != nil {
			log.Printf("healthz server error: %v", err)
			cancel()
		}
	}()

	if err := a.StartWorker(ctx); err != nil {
		cancel()
		return err
	}

	a.Wait()
	log.Println("✅ app shut down cleanly")
	return nil
}
