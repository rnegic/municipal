package service

import (
	"context"
	"log/slog"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

// IncidentRow is the service-level view of an incident (transport maps it to the wire format;
// it never sees repository.IncidentRow or gen/db types directly).
type IncidentRow struct {
	ID            int64
	HouseID       int64
	Title         string
	Description   string
	Severity      string
	Status        string
	AffectedCount int
	CreatedAt     time.Time
	JoinedByMe    bool
	ConfirmedByMe bool
}

func toIncidentRow(r repository.IncidentRow) IncidentRow {
	return IncidentRow{
		ID: r.ID, HouseID: r.HouseID, Title: r.Title, Description: r.Description,
		Severity: r.Severity, Status: r.Status, AffectedCount: r.Subscribers,
		CreatedAt: r.CreatedAt, JoinedByMe: r.JoinedByMe, ConfirmedByMe: r.ConfirmedByMe,
	}
}

// CreateIncident joins an open duplicate (same house+title+riser, opened within
// domain.DedupWindow) instead of creating a new one; created=false signals the caller to
// respond 200 (joined existing) instead of 201 (new).
func (s *Service) CreateIncident(ctx context.Context, houseID, reporterID int64, title, description, severityStr string, entrance, riser *string) (row IncidentRow, created bool, err error) {
	sev := domain.Severity(severityStr)
	if !sev.Valid() || title == "" || description == "" {
		return IncidentRow{}, false, ErrInvalidInput
	}
	open, err := s.repo.OpenIncidents(ctx, houseID)
	if err != nil {
		return IncidentRow{}, false, err
	}
	if dup := domain.FindDuplicate(houseID, title, riser, time.Now(), open); dup != 0 {
		if err := s.repo.Subscribe(ctx, dup, reporterID); err != nil {
			return IncidentRow{}, false, err
		}
		r, err := s.repo.GetIncident(ctx, dup, reporterID)
		return toIncidentRow(r), false, err
	}
	id, err := s.repo.CreateIncident(ctx, houseID, reporterID, title, description, sev, entrance, riser)
	if err != nil {
		return IncidentRow{}, false, err
	}
	if err := s.syncUnregistered(ctx); err != nil {
		slog.Warn("uk register deferred to worker", "incident", id, "err", err)
	}
	r, err := s.repo.GetIncident(ctx, id, reporterID)
	return toIncidentRow(r), true, err
}

func (s *Service) GetIncident(ctx context.Context, id, userID int64) (IncidentRow, error) {
	r, err := s.repo.GetIncident(ctx, id, userID)
	return toIncidentRow(r), err
}

func (s *Service) ListActiveIncidents(ctx context.Context, houseID, userID int64) ([]IncidentRow, error) {
	rows, err := s.repo.ListActiveIncidents(ctx, houseID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]IncidentRow, len(rows))
	for i, r := range rows {
		out[i] = toIncidentRow(r)
	}
	return out, nil
}

// ListRequests returns the reporter's own incidents in the house, paginated.
func (s *Service) ListRequests(ctx context.Context, houseID, reporterID, offset, limit int64) ([]IncidentRow, int64, error) {
	rows, total, err := s.repo.ListReporterIncidents(ctx, houseID, reporterID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]IncidentRow, len(rows))
	for i, r := range rows {
		out[i] = toIncidentRow(r)
	}
	return out, total, nil
}

// JoinIncident ("у меня тоже") is idempotent: joined is always true on success.
func (s *Service) JoinIncident(ctx context.Context, incidentID, userID int64) (affectedCount int, joined bool, err error) {
	if err := s.repo.Subscribe(ctx, incidentID, userID); err != nil {
		return 0, false, err
	}
	n, err := s.repo.SubscriberCount(ctx, incidentID)
	return n, true, err
}
