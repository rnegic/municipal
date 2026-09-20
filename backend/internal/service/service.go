// Package service — бизнес-оркестрация: применяет domain-правила поверх repository и внешних
// клиентов (MAX, DaData). Ничего не знает про HTTP/JSON — ни oapi, ни net/http сюда не попадают.
package service

import (
	"errors"

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
	repo *repository.Store
	maxc *maxclient.Client
	dd   *dadata.Client
}

func New(repo *repository.Store, maxc *maxclient.Client, dd *dadata.Client) *Service {
	return &Service{repo: repo, maxc: maxc, dd: dd}
}
