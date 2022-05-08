package main

import (
	"context"
	"fmt"
	"github.com/ATenderholt/rainbow/settings"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type MotoService interface {
	ReplayAllRequests(ctx context.Context) error
}

type App struct {
	cfg  *settings.Config
	srv  *http.Server
	moto MotoService
}

func NewApp(cfg *settings.Config, mux *chi.Mux, moto MotoService) App {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BasePort),
		Handler: mux,
	}

	return App{
		cfg:  cfg,
		srv:  srv,
		moto: moto,
	}
}

func (app App) Start() error {
	errors := make(chan error, 5)
	go app.StartHttp(errors)
	go app.StartMoto(errors)

	return nil
}

func (app App) StartHttp(errors chan error) {
	logger.Infof("Starting HTTP server on port %d", app.cfg.BasePort)
	err := app.srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		logger.Errorf("Problem starting HTTP server: %v", err)
		errors <- err
	}
}

func (app App) StartMoto(errors chan error) {
	err := app.moto.ReplayAllRequests(context.Background())
	if err != nil {
		errors <- err
	}
}

func (app App) Shutdown() error {
	return nil
}
