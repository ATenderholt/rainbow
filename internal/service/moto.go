package service

import (
	"context"
	"fmt"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-rel/rel"
)

type MotoService struct {
	repo       rel.Repository
	predicates map[string]persistPredicate
}

func NewMotoService(repo rel.Repository) MotoService {
	predicates := make(map[string]persistPredicate)
	predicates["iam"] = persistIamRequest
	predicates["ssm"] = persistSsmRequest

	return MotoService{
		repo:       repo,
		predicates: predicates,
	}
}

func (moto MotoService) shouldPersist(request domain.MotoRequest) bool {
	predicate, ok := moto.predicates[request.Service]
	if ok {
		return predicate(request)
	}

	return true
}

func (moto MotoService) SaveRequest(ctx context.Context, request domain.MotoRequest) error {
	if !moto.shouldPersist(request) {
		logger.Infof("NOT saving Moto request: %+v", request)
		return nil
	}

	logger.Infof("Saving Moto request: %+v", request)
	err := moto.repo.Insert(ctx, &request)
	if err != nil {
		e := fmt.Errorf("unable to insert moto request %+v: %v", request, err)
		logger.Error(e)
		return e
	}

	return nil
}
