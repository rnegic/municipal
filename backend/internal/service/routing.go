package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"ukapp/internal/domain"
)

type Classifier interface {
	Classify(ctx context.Context, text string) (domain.Prediction, error)
}

var classifyTimeout = 10 * time.Second

type Analysis struct {
	Category  domain.Category
	Authority domain.Authority
	Reasoning string
}

func (s *Service) WithClassifier(c Classifier, threshold float64) *Service {
	s.cls, s.threshold = c, threshold
	return s
}

func (s *Service) classify(ctx context.Context, text string) (domain.Prediction, error) {
	if s.cls == nil {
		return domain.Prediction{}, ErrModelUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()
	select {
	case s.inflight <- struct{}{}:
		defer func() { <-s.inflight }()
	case <-ctx.Done():
		return domain.Prediction{}, ErrModelUnavailable
	}
	start := time.Now()
	p, err := s.cls.Classify(ctx, text)
	sum := sha256.Sum256([]byte(text))
	slog.Info("classify", "description_hash", hex.EncodeToString(sum[:8]), "category", p.Category, "p", p.P,
		"confident", err == nil && p.P >= s.threshold, "latency_ms", time.Since(start).Milliseconds(), "err", err)
	if err != nil {
		return domain.Prediction{}, fmt.Errorf("%w: %v", ErrModelUnavailable, err)
	}
	return p, nil
}

func (s *Service) AnalyzeIncident(ctx context.Context, description string) (Analysis, error) {
	d, ok := domain.ValidDescription(description)
	if !ok {
		return Analysis{}, ErrInvalidInput
	}
	p, err := s.classify(ctx, d)
	if err != nil {
		return Analysis{}, err
	}
	if p.P < s.threshold {
		return Analysis{}, ErrModelUnavailable
	}
	auth := p.Category.Authority()
	return Analysis{Category: p.Category, Authority: auth, Reasoning: auth.Reasoning()}, nil
}
