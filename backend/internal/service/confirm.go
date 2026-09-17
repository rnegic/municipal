package service

import (
	"context"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

// ConfirmIncident records the resident's "почикили" confirmation. Only allowed while the
// incident is verifying (ErrInvalidStatus otherwise). Once enough residents confirmed
// (domain.ShouldClose), the incident is atomically closed and subscribers notified.
func (s *Service) ConfirmIncident(ctx context.Context, incidentID, userID int64) (domain.IncidentStatus, time.Time, error) {
	status, err := s.repo.FindIncidentStatus(ctx, incidentID)
	if err != nil {
		return "", time.Time{}, err
	}
	if status != domain.IncidentVerifying {
		return "", time.Time{}, ErrInvalidStatus
	}
	confirmedAt, err := s.repo.UpsertConfirmation(ctx, incidentID, userID)
	if err != nil {
		return "", time.Time{}, err
	}
	subs, err := s.repo.SubscriberCount(ctx, incidentID)
	if err != nil {
		return "", time.Time{}, err
	}
	cnt, err := s.repo.ConfirmationCount(ctx, incidentID)
	if err != nil {
		return "", time.Time{}, err
	}
	if !domain.ShouldClose(subs, cnt) {
		return domain.IncidentVerifying, confirmedAt, nil
	}
	payload := repository.OutboxPayload{Text: "Ваша проблема закрыта: жители подтвердили, что всё работает."}
	closed, err := s.repo.CloseIfVerifying(ctx, incidentID, "incident_done", payload)
	if err != nil {
		return "", time.Time{}, err
	}
	if !closed {
		return domain.IncidentVerifying, confirmedAt, nil
	}
	return domain.IncidentDone, confirmedAt, nil
}
