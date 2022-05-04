package http

import (
	"context"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MotoService interface {
	SaveRequest(ctx context.Context, request domain.MotoRequest) error
}

type SqsService interface {
	ParseAction(string) string
	SaveAttributes(string) error
	DecorateAttributes(string, []byte) ([]byte, error)
}

func NewChiMux(moto MotoService, sqs SqsService, proxy Proxy) *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(extractService)
	mux.Use(motoMiddleware(moto))
	mux.Use(sqsMiddleware(sqs))

	mux.Head("/*", proxy.handle)
	mux.Get("/*", proxy.handle)
	mux.Post("/*", proxy.handle)
	mux.Put("/*", proxy.handle)
	mux.Delete("/*", proxy.handle)

	return mux
}
