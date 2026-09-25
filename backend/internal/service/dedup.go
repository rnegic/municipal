package service

import (
	"context"
	"log/slog"
	"time"

	"ukapp/internal/domain"
)

type Matcher interface {
	Match(ctx context.Context, req domain.OpenIncident, cands []domain.OpenIncident) (domain.DedupMatch, error)
}

func (s *Service) WithMatcher(m Matcher) *Service {
	s.matcher = m
	return s
}

func (s *Service) findDuplicate(ctx context.Context, req domain.OpenIncident, open []domain.OpenIncident) (int64, string) {
	now := time.Now()
	if s.matcher != nil {
		if cands := domain.DedupCandidates(open, now); len(cands) > 0 {
			m, err := s.matcher.Match(ctx, req, cands)
			if err == nil {
				slog.Info("dedup", "house", req.HouseID, "candidates", len(cands), "match", m.IncidentID, "p", m.P, "model_version", m.ModelVersion)
				return m.IncidentID, m.ModelVersion
			}
			slog.Warn("dedup model failed, using rule", "house", req.HouseID, "err", err)
		}
	}
	return domain.FindDuplicate(req.HouseID, req.Category, req.Riser, now, open), domain.DedupVersion
}
