package http

import (
	"context"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io/ioutil"
	"net/http"
)

type MotoService interface {
	SaveRequest(ctx context.Context, request domain.MotoRequest) error
}

func handler(w http.ResponseWriter, r *http.Request) {
	logger.Infof("Handling %s for service %s", r.URL.Path, ServiceFromRequest(r))
	payload, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	logger.Infof("Got payload: %s", string(payload))
}

func NewChiMux(service MotoService) *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(extractService)
	mux.Use(motoMiddleware(service))

	mux.Head("/*", handler)
	mux.Get("/*", handler)
	mux.Post("/*", handler)
	mux.Put("/*", handler)
	mux.Delete("/*", handler)

	return mux
}
