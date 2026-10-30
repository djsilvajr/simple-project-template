package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/database"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/mailer"
	"github.com/djsilvajr/simple-project-template/back/internal/router"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Connect(cfg)
	if err != nil {
		return err
	}
	defer database.Close(db)

	handler, err := router.Setup(router.Dependencies{
		Config: cfg,
		DB:     db,
		Mailer: mailer.New(cfg.Mail),
	})
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("api: %s (%s) ouvindo em %s", cfg.App.Name, cfg.App.Env, server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Println("api: encerrando...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}
