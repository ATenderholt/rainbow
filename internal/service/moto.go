package service

import (
	"context"
	"fmt"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-rel/rel"
)

type MotoService struct {
	repo rel.Repository
}

func NewMotoService(repo rel.Repository) MotoService {
	return MotoService{
		repo: repo,
	}
}

func (service MotoService) SaveRequest(ctx context.Context, request domain.MotoRequest) error {
	logger.Infof("Saving Moto request: %+v", request)
	err := service.repo.Insert(ctx, &request)
	if err != nil {
		e := fmt.Errorf("unable to insert moto request %+v: %v", request, err)
		logger.Error(e)
		return e
	}

	return nil
}
