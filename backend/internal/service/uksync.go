package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

const ukRegisterBatch = 50

func (s *Service) syncUnregistered(ctx context.Context) error {
	rows, err := s.repo.UnregisteredIncidents(ctx, ukRegisterBatch)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.HouseExternalID == nil {
			slog.Warn("uk sync: house without external id, skip", "incident", r.ID)
			continue
		}
		id, _, err := s.uk.RegisterIncident(ctx, UkIncident{
			ExternalRef: strconv.FormatInt(r.ID, 10), HouseID: *r.HouseExternalID,
			Title: r.Title, Description: r.Description, Severity: r.Severity,
			Entrance: r.Entrance, Riser: r.Riser, Suspicious: r.Suspicious,
		})
		if err != nil {
			return fmt.Errorf("register incident %d: %w", r.ID, err)
		}
		if err := s.repo.SetIncidentExternalID(ctx, r.ID, id); err != nil {
			return err
		}
	}
	return nil
}

const ukSyncDefaultTick = 5 * time.Second

var ukStatusText = map[domain.IncidentStatus]string{
	domain.IncidentAccepted:   "Управляющая компания приняла заявку.",
	domain.IncidentInProgress: "Проблема взята в работу управляющей компанией.",
	domain.IncidentVerifying:  "УК сообщает, что проблема решена. Подтвердите, пожалуйста, в приложении.",
	domain.IncidentDone:       "Проблема закрыта управляющей компанией.",
	domain.IncidentFalseAlarm: "Управляющая компания закрыла заявку с отметкой «Ложный вызов».",
}

func (s *Service) RunUkSyncWorker(ctx context.Context) {
	tick := ukSyncDefaultTick
	if d, err := time.ParseDuration(os.Getenv("UK_SYNC_INTERVAL")); err == nil && d > 0 {
		tick = d
	}
	s.ukSince = time.Now().Add(-24 * time.Hour)
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.syncUnregistered(ctx); err != nil {
				slog.Warn("uk sync: register", "err", err)
			}
			if err := s.syncStatuses(ctx); err != nil {
				slog.Warn("uk sync: statuses", "err", err)
			}
		}
	}
}

func (s *Service) syncStatuses(ctx context.Context) error {
	ups, err := s.uk.IncidentUpdates(ctx, s.ukSince)
	if err != nil {
		return err
	}
	maxSeen := s.ukSince
	for _, u := range ups {
		if u.UpdatedAt.After(maxSeen) {
			maxSeen = u.UpdatedAt
		}
		text, ok := ukStatusText[u.Status]
		if !ok {
			continue
		}
		if _, err := s.repo.ApplyUkStatus(ctx, u.ID, u.Status, "uk_status_"+string(u.Status), repository.OutboxPayload{Text: text}); err != nil {
			return fmt.Errorf("apply status %s for %s: %w", u.Status, u.ID, err)
		}
	}
	s.ukSince = maxSeen
	return nil
}
