//go:build wireinject
// +build wireinject

package main

import (
	"github.com/ATenderholt/rainbow/settings"
	"github.com/google/wire"
)

func NewApp(cfg *settings.Config) App {
	return App{
		cfg: cfg,
	}
}

func InjectApp(cfg *settings.Config) (App, error) {
	wire.Build(
		NewApp,
	)

	return App{}, nil
}
