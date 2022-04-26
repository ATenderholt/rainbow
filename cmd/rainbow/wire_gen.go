// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/ATenderholt/rainbow/settings"
)

// Injectors from inject.go:

func InjectApp(cfg *settings.Config) (App, error) {
	app := NewApp(cfg)
	return app, nil
}

// inject.go:

func NewApp(cfg *settings.Config) App {
	return App{
		cfg: cfg,
	}
}
