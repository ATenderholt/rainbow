package service

import (
	"context"
	"fmt"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/ATenderholt/rainbow/settings"
	"github.com/go-rel/rel"
	"net/http"
	"strings"
	"time"
)

type MotoService struct {
	cfg        *settings.Config
	repo       rel.Repository
	predicates map[string]persistPredicate
}

func NewMotoService(cfg *settings.Config, repo rel.Repository) MotoService {
	predicates := make(map[string]persistPredicate)
	predicates["iam"] = persistIamRequest
	predicates["ssm"] = persistSsmRequest

	return MotoService{
		cfg:        cfg,
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

func (moto MotoService) ReplayAllRequests(ctx context.Context) error {
	requests, err := moto.findAllRequests(ctx)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof("Found %d Moto requests to replay", len(requests))
	for i, request := range requests {
		err = moto.replayToMoto(ctx, request)
		if err != nil {
			e := fmt.Errorf("unable to replay moto request %d: %v", i, err)
			logger.Error(e)
			return e
		}
	}

	logger.Infof("Done replaying all moto requests")

	return nil
}

func (moto MotoService) findAllRequests(ctx context.Context) ([]domain.MotoRequest, error) {
	logger.Infof("Finding all Moto requests")

	var requests []domain.MotoRequest
	err := moto.repo.FindAll(ctx, &requests)
	if err != nil {
		e := fmt.Errorf("unable to find all moto requests: %v", err)
		logger.Error(e)
		return nil, e
	}

	return requests, err
}

func (moto MotoService) replayToMoto(ctx context.Context, request domain.MotoRequest) error {
	logger.Infof("Replaying %s request #%d to moto ...", request.Service, request.ID)

	url := "http://" + moto.cfg.MotoHost() + "/" + request.Path

	proxyReq, _ := http.NewRequest(request.Method, url, strings.NewReader(request.Payload))
	proxyReq.Header.Set("Content-Type", request.ContentType)
	proxyReq.Header.Set("Authorization", request.Authorization)
	if len(request.Target) > 0 {
		proxyReq.Header.Set("X-Amz-Target", request.Target)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	proxyReq = proxyReq.Clone(timeoutCtx)
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		e := fmt.Errorf("unable to replay request %+v to moto: %v", request, err)
		logger.Error(e)
		return e
	}

	logger.Infof("Got following repsonse from Moto: %+v", resp)
	return nil
}
