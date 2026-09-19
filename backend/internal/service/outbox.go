package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"ukapp/internal/repository"
)

const (
	outboxTick        = 2 * time.Second
	outboxBatch       = 25 // < 30 rps limit of MAX Bot API
	outboxMaxAttempts = 5
)

// ponytail: single in-process worker, no leader election. Run one api replica or accept duplicate sends.
func (s *Service) RunOutboxWorker(ctx context.Context) {
	t := time.NewTicker(outboxTick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.drainOutbox(ctx); err != nil {
				slog.Error("outbox", "err", err)
			}
		}
	}
}

func (s *Service) drainOutbox(ctx context.Context) error {
	items, err := s.repo.FetchPendingOutbox(ctx, outboxBatch)
	if err != nil {
		return err
	}
	for _, it := range items {
		var p repository.OutboxPayload
		_ = json.Unmarshal([]byte(it.Payload), &p)
		if err := s.maxc.SendMessage(ctx, it.TargetMaxUserID, p.Text); err != nil {
			uerr := s.repo.MarkOutboxAttemptFailed(ctx, it.ID, it.Attempts, outboxMaxAttempts)
			slog.Warn("outbox send failed", "id", it.ID, "attempt", it.Attempts+1, "err", err, "update_err", uerr)
			continue
		}
		if err := s.repo.MarkOutboxSent(ctx, it.ID, it.Attempts+1); err != nil {
			return err
		}
	}
	return nil
}
