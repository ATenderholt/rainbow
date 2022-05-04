//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"
	"github.com/ATenderholt/rainbow/internal/http"
	"github.com/ATenderholt/rainbow/internal/service"
	"github.com/ATenderholt/rainbow/settings"
	"github.com/go-rel/rel"
	"github.com/go-rel/sqlite3"
	"github.com/google/wire"
)

func InjectDb(db *sql.DB) rel.Repository {
	adapter := sqlite3.New(db)
	return rel.New(adapter)
}

func InjectApp(cfg *settings.Config, db *sql.DB) (App, error) {
	wire.Build(
		NewApp,
		InjectDb,
		http.NewChiMux,
		http.NewProxy,
		service.NewMotoService,
		service.NewSqsService,
		wire.Bind(new(http.MotoService), new(service.MotoService)),
		wire.Bind(new(http.ElasticService), new(service.SqsService)),
	)

	return App{}, nil
}
