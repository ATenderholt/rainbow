package service

import (
	"github.com/go-rel/rel"
)

type MotoService struct {
	repo rel.Repository
}

func NewMotoService(repo rel.Repository) *MotoService {
	return &MotoService{
		repo: repo,
	}
}
