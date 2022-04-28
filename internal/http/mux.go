package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewChiMux(router *ServiceRouter) *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)

	mux.Head("/*", router.RouteByAuthorization)
	mux.Get("/*", router.RouteByAuthorization)
	mux.Post("/*", router.RouteByAuthorization)
	mux.Put("/*", router.RouteByAuthorization)
	mux.Delete("/*", router.RouteByAuthorization)

	return mux
}
