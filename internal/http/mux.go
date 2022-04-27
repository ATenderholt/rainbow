package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func handle(w http.ResponseWriter, r *http.Request) {
	logger.Infof("Handling %s", r.URL.Path)
}

func NewChiMux() *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)

	mux.Post("/*", handle)
	return mux
}
