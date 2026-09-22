package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

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

func (s *Service) CreateIncident(ctx context.Context, houseID, reporterID int64, title, description, severityStr string, entrance, riser *string) (row IncidentRow, created bool, err error) {
	sev := domain.Severity(severityStr)
	if !sev.Valid() || title == "" || description == "" {
		return IncidentRow{}, false, ErrInvalidInput
	}
	open, err := s.repo.OpenIncidents(ctx, houseID)
	if err != nil {
		return IncidentRow{}, false, err
	}
	report := repository.ReportInput{
		ReporterID: reporterID, HouseID: houseID, Title: title, Description: description,
		Severity: sev, Entrance: entrance, Riser: riser, DedupVersion: domain.DedupVersion,
	}
	if dup := domain.FindDuplicate(houseID, title, riser, time.Now(), open); dup != 0 {
		if err := s.repo.Subscribe(ctx, dup, reporterID); err != nil {
			return IncidentRow{}, false, err
		}
		report.IncidentID, report.Outcome = dup, domain.ReportJoined
		s.saveReport(ctx, report)
		row, err = s.GetIncident(ctx, dup, reporterID)
		return row, false, err
	}
	id, err := s.repo.CreateIncident(ctx, houseID, reporterID, title, description, sev, entrance, riser, domain.SLA(sev))
	if err != nil {
		return IncidentRow{}, false, err
	}
	report.IncidentID, report.Outcome = id, domain.ReportCreated
	s.saveReport(ctx, report)
	if err := s.syncUnregistered(ctx); err != nil {
		slog.Warn("uk register deferred to worker", "incident", id, "err", err)
	}
	row, err = s.GetIncident(ctx, id, reporterID)
	return row, true, err
}

func (s *Service) saveReport(ctx context.Context, r repository.ReportInput) {
	if err := s.repo.AddReport(ctx, r); err != nil {
		slog.Warn("incident report not saved", "incident", r.IncidentID, "outcome", r.Outcome, "err", err)
	}
}

func (s *Service) ListActiveIncidents(ctx context.Context, houseID, userID int64) ([]IncidentRow, error) {
	rows, err := s.repo.ListActiveIncidents(ctx, houseID, userID)
	if err != nil {
		return nil, err
	}
	return s.toIncidentRows(ctx, rows)
}

func (s *Service) ListRequests(ctx context.Context, houseID, userID, offset, limit int64) ([]IncidentRow, int64, error) {
	rows, total, err := s.repo.ListUserIncidents(ctx, houseID, userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out, err := s.toIncidentRows(ctx, rows)
	return out, total, err
}

func (s *Service) JoinIncident(ctx context.Context, incidentID, userID int64) (affectedCount int, joined bool, err error) {
	if err := s.repo.Subscribe(ctx, incidentID, userID); err != nil {
		return 0, false, err
	}
	n, err := s.repo.SubscriberCount(ctx, incidentID)
	return n, true, err
}

func (s *Service) deepLink(incidentID int64) string {
	return fmt.Sprintf("https://max.ru/%s?startapp=inc_%d", s.botName, incidentID)
}

func (s *Service) SetIncidentStatus(ctx context.Context, incidentID, userID int64, to domain.IncidentStatus) (IncidentRow, error) {
	from, ok := domain.DispatcherTransition(to)
	if !ok {
		return IncidentRow{}, ErrInvalidStatus
	}
	text := ukStatusText[to] + " " + s.deepLink(incidentID)
	moved, err := s.repo.TransitionIncident(ctx, incidentID, from, to, "incident_"+string(to), repository.OutboxPayload{Text: text})
	if err != nil {
		return IncidentRow{}, err
	}
	if !moved {
		return IncidentRow{}, ErrInvalidStatus
	}
	return s.GetIncident(ctx, incidentID, userID)
}
