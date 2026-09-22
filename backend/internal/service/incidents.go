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
	Entrance      *string
	Riser         *string
	Severity      string
	Status        string
	AffectedCount int
	CreatedAt     time.Time
	DueAt         *time.Time
	JoinedByMe    bool
	ConfirmedByMe bool
	PhotoIDs      []int64
}

// toIncidentRows maps repository rows and attaches photo ids in one extra query.
func (s *Service) toIncidentRows(ctx context.Context, rs []repository.IncidentRow) ([]IncidentRow, error) {
	ids := make([]int64, len(rs))
	for i, r := range rs {
		ids[i] = r.ID
	}
	photos, err := s.repo.PhotoIDsByIncident(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]IncidentRow, len(rs))
	for i, r := range rs {
		out[i] = IncidentRow{
			ID: r.ID, HouseID: r.HouseID, Title: r.Title, Description: r.Description, Entrance: r.Entrance, Riser: r.Riser,
			Severity: r.Severity, Status: r.Status, AffectedCount: r.Subscribers,
			CreatedAt: r.CreatedAt, DueAt: r.DueAt, JoinedByMe: r.JoinedByMe, ConfirmedByMe: r.ConfirmedByMe,
			PhotoIDs: photos[r.ID],
		}
	}
	return out, nil
}

func (s *Service) GetIncident(ctx context.Context, id, userID int64) (IncidentRow, error) {
	r, err := s.repo.GetIncident(ctx, id, userID)
	if err != nil {
		return IncidentRow{}, err
	}
	rows, err := s.toIncidentRows(ctx, []repository.IncidentRow{r})
	if err != nil {
		return IncidentRow{}, err
	}
	return rows[0], nil
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
		row, err = s.GetIncident(ctx, dup, reporterID)
		return row, false, err
	}
	id, err := s.repo.CreateIncident(ctx, houseID, reporterID, title, description, sev, entrance, riser, domain.SLA(sev))
	if err != nil {
		return IncidentRow{}, false, err
	}
	if err := s.syncUnregistered(ctx); err != nil {
		slog.Warn("uk register deferred to worker", "incident", id, "err", err)
	}
	row, err = s.GetIncident(ctx, id, reporterID)
	return row, true, err
}

func (s *Service) ListActiveIncidents(ctx context.Context, houseID, userID int64) ([]IncidentRow, error) {
	rows, err := s.repo.ListActiveIncidents(ctx, houseID, userID)
	if err != nil {
		return nil, err
	}
	return s.toIncidentRows(ctx, rows)
}

// ListRequests returns the incidents the user reported or joined in the house, paginated.
func (s *Service) ListRequests(ctx context.Context, houseID, userID, offset, limit int64) ([]IncidentRow, int64, error) {
	rows, total, err := s.repo.ListUserIncidents(ctx, houseID, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out, err := s.toIncidentRows(ctx, rows)
	return out, total, err
}

// JoinIncident ("у меня тоже") is idempotent: joined is always true on success.
func (s *Service) JoinIncident(ctx context.Context, incidentID, userID int64) (affectedCount int, joined bool, err error) {
	if err := s.repo.Subscribe(ctx, incidentID, userID); err != nil {
		return 0, false, err
	}
	n, err := s.repo.SubscriberCount(ctx, incidentID)
	return n, true, err
}
