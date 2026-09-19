package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
)

const ukRegisterBatch = 50

// syncUnregistered registers incidents the UK system doesn't know yet (external_id IS NULL).
// Called right after CreateIncident (best effort) and by RunUkSyncWorker (retry).
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
			Entrance: r.Entrance, Riser: r.Riser,
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
