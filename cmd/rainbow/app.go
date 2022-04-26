package main

import "github.com/ATenderholt/rainbow/settings"

type App struct {
	cfg *settings.Config
}

func (app App) Start() error {
	return nil
}

func (app App) Shutdown() error {
	return nil
}
