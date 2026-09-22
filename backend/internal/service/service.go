package service

import (
	"errors"
	"time"

	"ukapp/internal/dadata"
	"ukapp/internal/maxclient"
	"ukapp/internal/repository"
)

var (
	ErrNotFound      = repository.ErrNotFound
	ErrInvalidInput  = errors.New("invalid input")
	ErrInvalidStatus = errors.New("status transition not allowed")
	ErrForbidden     = errors.New("forbidden")
	ErrRateLimited   = errors.New("rate limited")
)

type Service struct {
	repo    *repository.Store
	maxc    *maxclient.Client
	dd      *dadata.Client
	uk      UkProvider
	botName string
	ukSince time.Time
}

func New(repo *repository.Store, maxc *maxclient.Client, dd *dadata.Client, uk UkProvider, botName string) *Service {
	return &Service{repo: repo, maxc: maxc, dd: dd, uk: uk, botName: botName}
}
