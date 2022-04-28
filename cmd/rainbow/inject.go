//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ATenderholt/rainbow/internal/http"
	"github.com/ATenderholt/rainbow/settings"
	"github.com/google/wire"
)

var api = wire.NewSet(
	http.NewChiMux,
	http.NewServiceRouter,
)

func InjectApp(cfg *settings.Config) (App, error) {
	wire.Build(
		NewApp,
		api,
	)

	return App{}, nil
}
