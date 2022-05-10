package main

import (
	"context"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow/logging"
	"github.com/ATenderholt/rainbow/settings"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type MotoService interface {
	ReplayAllRequests(ctx context.Context) error
	Start(ctx context.Context) error
}

type SqsService interface {
	Start(ctx context.Context) error
}

type App struct {
	cfg    *settings.Config
	docker *dockerlib.DockerController
	srv    *http.Server
	moto   MotoService
	sqs    SqsService
}

func NewApp(cfg *settings.Config, docker *dockerlib.DockerController, mux *chi.Mux, moto MotoService, sqs SqsService) App {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BasePort),
		Handler: mux,
	}

	return App{
		cfg:    cfg,
		docker: docker,
		srv:    srv,
		moto:   moto,
		sqs:    sqs,
	}
}

func (app App) Start() error {
	dockerlib.SetLogger(logging.NewLogger().Desugar().Named("dockerlib"))

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		s := <-c
		logger.Infof("Received signal %v", s)
		app.Shutdown()
	}()

	err := app.StartMoto()
	if err != nil {
		e := fmt.Errorf("unable to start moto: %v", err)
		logger.Error(e)
		return e
	}

	err = app.StartSqs()
	if err != nil {
		e := fmt.Errorf("unable to start sqs: %v", err)
		logger.Error(e)
		return e
	}
	
	return app.StartHttp()
}

func (app App) StartHttp() error {
	logger.Infof("Starting HTTP server on port %d", app.cfg.BasePort)
	err := app.srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		e := fmt.Errorf("unable to start HTTP server: %v", err)
		logger.Error(e)
		return e
	}

	return nil
}

func (app App) StartMoto() error {
	ctx := context.Background()

	err := app.moto.Start(ctx)
	if err != nil {
		logger.Errorf("unable to start moto: %v", err)
		return err
	}

	err = app.moto.ReplayAllRequests(ctx)
	if err != nil {
		logger.Errorf("unable to replay Moto requests: %v", err)
		return err
	}

	return nil
}

func (app App) StartSqs() error {
	ctx := context.Background()
	err := app.sqs.Start(ctx)
	if err != nil {
		logger.Errorf("unable to start sqs: %v", err)
		return err
	}

	return nil
}

func (app App) Shutdown() error {
	logger.Info("Starting shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := app.docker.ShutdownAll(ctx)
	if err != nil {
		logger.Errorf("unable to shutdown all docker containers: %v", err)
	}

	err = app.srv.Shutdown(ctx)
	if err != nil {
		e := fmt.Errorf("error during shutdown: %v", err)
		logger.Error(e)
		return e
	}

	logger.Infof("Shutdown complete")
	return nil
}
