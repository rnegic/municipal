package service

import (
	"errors"
	"sync"
	"time"

	"ukapp/internal/dadata"
	"ukapp/internal/maxclient"
	"ukapp/internal/repository"
	"ukapp/internal/ukauth"
)

var (
	ErrNotFound         = repository.ErrNotFound
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidStatus    = errors.New("status transition not allowed")
	ErrForbidden        = errors.New("forbidden")
	ErrRateLimited      = errors.New("rate limited")
	ErrModelUnavailable = errors.New("model unavailable")
)

type Service struct {
	repo    *repository.Store
	maxc    *maxclient.Client
	dd      *dadata.Client
	uk      UkProvider
	botName string
	ukSince time.Time
	tokens  *ukauth.TokenSigner

	cls       Classifier
	threshold float64
	matcher   Matcher
	orgs      OrgDirectory
	inflight  chan struct{}

	houseLocks sync.Map
}

func New(repo *repository.Store, maxc *maxclient.Client, dd *dadata.Client, uk UkProvider, botName string) *Service {
	tokens, err := ukauth.NewTokenSigner("")
	if err != nil {
		panic(err)
	}
	return &Service{repo: repo, maxc: maxc, dd: dd, uk: uk, botName: botName, tokens: tokens, inflight: make(chan struct{}, 2)}
}

func (s *Service) lockHouse(houseID int64) (unlock func()) {
	m, _ := s.houseLocks.LoadOrStore(houseID, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}
