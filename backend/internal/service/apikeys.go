package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"ukapp/internal/ukauth"

	"ukapp/gen/db/ukapp/public/model"
)

type CreatedApiKey struct {
	Key   model.UkAPIKey
	Plain string
}

func (s *Service) CreateApiKey(ctx context.Context, u model.AppUser, name string) (CreatedApiKey, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return CreatedApiKey{}, ErrInvalidInput
	}
	plain, prefix, hash, err := ukauth.NewAPIKey()
	if err != nil {
		return CreatedApiKey{}, err
	}
	k, err := s.repo.CreateApiKey(ctx, *u.UkID, u.ID, name, prefix, hash)
	if err != nil {
		return CreatedApiKey{}, err
	}
	return CreatedApiKey{Key: k, Plain: plain}, nil
}

func (s *Service) ListApiKeys(ctx context.Context, u model.AppUser) ([]model.UkAPIKey, error) {
	return s.repo.ListApiKeys(ctx, *u.UkID)
}

func (s *Service) RevokeApiKey(ctx context.Context, u model.AppUser, id int64) error {
	return s.repo.RevokeApiKey(ctx, *u.UkID, id)
}

func (s *Service) AuthenticateApiKey(ctx context.Context, raw string) (model.AppUser, error) {
	k, err := s.repo.ApiKeyByHash(ctx, ukauth.HashAPIKey(raw))
	if errors.Is(err, ErrNotFound) {
		return model.AppUser{}, ErrForbidden
	}
	if err != nil {
		return model.AppUser{}, err
	}
	now := time.Now()
	if !s.keyLimiter.allow(k.ID, now) {
		return model.AppUser{}, ErrRateLimited
	}
	if k.LastUsedAt == nil || now.Sub(*k.LastUsedAt) > time.Minute {
		if err := s.repo.TouchApiKey(ctx, k.ID); err != nil {
			return model.AppUser{}, err
		}
	}
	return s.repo.GetUser(ctx, k.UserID)
}

type keyLimiter struct {
	mu     sync.Mutex
	rps    int
	window map[int64]int64
	count  map[int64]int
}

func newKeyLimiter(rps int) *keyLimiter {
	return &keyLimiter{rps: rps, window: map[int64]int64{}, count: map[int64]int{}}
}

func (l *keyLimiter) allow(id int64, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	sec := now.Unix()
	if l.window[id] != sec {
		l.window[id], l.count[id] = sec, 0
	}
	l.count[id]++
	return l.count[id] <= l.rps
}
