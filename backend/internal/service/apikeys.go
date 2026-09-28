package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"ukapp/internal/domain"
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
	slog.Info("uk api key issued", "uk", *u.UkID, "key", k.ID, "prefix", k.Prefix, "by", u.ID)
	return CreatedApiKey{Key: k, Plain: plain}, nil
}

func (s *Service) ListApiKeys(ctx context.Context, u model.AppUser) ([]model.UkAPIKey, error) {
	return s.repo.ListApiKeys(ctx, *u.UkID)
}

func (s *Service) RevokeApiKey(ctx context.Context, u model.AppUser, id int64) error {
	if err := s.repo.RevokeApiKey(ctx, *u.UkID, id); err != nil {
		return err
	}
	slog.Info("uk api key revoked", "uk", *u.UkID, "key", id, "by", u.ID)
	return nil
}

func (s *Service) AuthenticateApiKey(ctx context.Context, raw string) (model.AppUser, error) {
	k, err := s.repo.ApiKeyByHash(ctx, ukauth.HashAPIKey(raw))
	if errors.Is(err, ErrNotFound) {
		return model.AppUser{}, ErrForbidden
	}
	if err != nil {
		return model.AppUser{}, err
	}
	org, err := s.repo.UkByID(ctx, k.UkID)
	if err != nil {
		return model.AppUser{}, err
	}
	now := time.Now()
	if !domain.LicenseActive(org.LicenseNumber, org.LicenseValidUntil, now) {
		return model.AppUser{}, ErrForbidden
	}
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

type keyWindow struct {
	sec int64
	n   int
}

type keyLimiter struct {
	mu   sync.Mutex
	rps  int
	keys map[int64]keyWindow
}

func newKeyLimiter(rps int) *keyLimiter {
	return &keyLimiter{rps: rps, keys: map[int64]keyWindow{}}
}

func (l *keyLimiter) allow(id int64, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.keys[id]
	if sec := now.Unix(); w.sec != sec {
		w = keyWindow{sec: sec}
	}
	w.n++
	l.keys[id] = w
	return w.n <= l.rps
}
