package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	logger.Infof("Handling %s for service %s", r.URL.Path, ServiceFromRequest(r))
}

func NewChiMux(router *ServiceRouter) *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(extractService)

	mux.Head("/*", handler)
	mux.Get("/*", handler)
	mux.Post("/*", handler)
	mux.Put("/*", handler)
	mux.Delete("/*", handler)

	return mux
}
