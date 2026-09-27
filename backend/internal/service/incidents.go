package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"ukapp/internal/domain"
	"ukapp/internal/repository"

	"ukapp/gen/db/ukapp/public/model"
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
	Category      *string
	MergedCount   int
	Supporters    []Supporter
}

type Supporter = repository.Supporter

const maxSupporters = 5

func (s *Service) toIncidentRows(ctx context.Context, rs []repository.IncidentRow) ([]IncidentRow, error) {
	ids := make([]int64, len(rs))
	for i, r := range rs {
		ids[i] = r.ID
	}
	photos, err := s.repo.PhotoIDsByIncident(ctx, ids)
	if err != nil {
		return nil, err
	}
	supporters, err := s.repo.SupportersByIncident(ctx, ids, maxSupporters)
	if err != nil {
		return nil, err
	}
	out := make([]IncidentRow, len(rs))
	for i, r := range rs {
		out[i] = IncidentRow{
			ID: r.ID, HouseID: r.HouseID, Title: r.Title, Description: r.Description, Entrance: r.Entrance, Riser: r.Riser,
			Severity: r.Severity, Status: r.Status, AffectedCount: r.Subscribers,
			CreatedAt: r.CreatedAt, DueAt: r.DueAt, JoinedByMe: r.JoinedByMe, ConfirmedByMe: r.ConfirmedByMe,
			PhotoIDs: photos[r.ID], Category: r.Category, MergedCount: int(r.MergedCount), Supporters: supporters[r.ID],
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

var (
	ErrPhotoRequired = errors.New("photo required")
	ErrPhotoNotOwned = repository.ErrPhotoNotOwned
)

type NotUKError struct{ Authority domain.Authority }

func (e NotUKError) Error() string { return "not uk responsibility: " + string(e.Authority) }

type NewIncident struct {
	Title               *string
	Description         string
	Category            domain.Category
	Entrance, FloorZone *string
	PhotoIDs            []int64
}

func tooLong(s *string, max int) bool { return s != nil && utf8.RuneCountInString(*s) > max }

func (s *Service) route(ctx context.Context, description string, category domain.Category) repository.Routing {
	r := repository.Routing{Source: "manual"}
	p, err := s.classify(ctx, description)
	if err != nil {
		return r
	}
	cat, conf := string(p.Category), float32(p.P)
	r.CategoryPredicted, r.Confidence = &cat, &conf
	if p.P >= s.threshold {
		r.Source = "auto"
	}
	if p.Category != category {
		slog.Info("category mismatch", "client", category, "predicted", p.Category, "p", p.P)
	}
	return r
}

func (s *Service) CreateIncident(ctx context.Context, houseID, reporterID int64, in NewIncident) (row IncidentRow, created bool, err error) {
	d, ok := domain.ValidDescription(in.Description)
	if !ok || !in.Category.Valid() || tooLong(in.Entrance, 40) || tooLong(in.FloorZone, 120) || len(in.PhotoIDs) > MaxPhotosPerIncident {
		return IncidentRow{}, false, ErrInvalidInput
	}
	// Название заявки задаёт пользователь; если он его не прислал, откатываемся
	// на название категории. На автоопределение категории название не влияет.
	title := in.Category.Title()
	if in.Title != nil {
		t, ok := domain.ValidTitle(*in.Title)
		if !ok {
			return IncidentRow{}, false, ErrInvalidInput
		}
		title = t
	}
	if auth := in.Category.Authority(); auth != domain.AuthorityUK {
		return IncidentRow{}, false, NotUKError{Authority: auth}
	}
	if in.Category.PhotoRequired() && len(in.PhotoIDs) == 0 {
		return IncidentRow{}, false, ErrPhotoRequired
	}
	defer s.lockHouse(houseID)()
	falseAlarms, recent, err := s.repo.ReporterStats(ctx, reporterID, time.Hour)
	if err != nil {
		return IncidentRow{}, false, err
	}
	if recent >= domain.MaxReportsPerHour {
		return IncidentRow{}, false, ErrRateLimited
	}
	routing := s.route(ctx, d, in.Category)
	slog.Info("incident routed", "house", houseID, "category", in.Category, "routing_source", routing.Source)
	open, err := s.repo.OpenIncidents(ctx, houseID, reporterID)
	if err != nil {
		return IncidentRow{}, false, err
	}
	sev := in.Category.Severity()
	req := domain.OpenIncident{
		HouseID: houseID, Category: in.Category, Title: title, Description: d,
		Entrance: in.Entrance, Riser: in.FloorZone, Severity: sev,
	}
	dup, dedupVersion := s.findDuplicate(ctx, req, open)
	report := repository.ReportInput{
		ReporterID: reporterID, HouseID: houseID, Title: title, Description: d,
		Severity: sev, Entrance: in.Entrance, Riser: in.FloorZone, DedupVersion: dedupVersion,
	}
	if dup != 0 {
		if err := s.repo.JoinWithPhotos(ctx, dup, reporterID, in.PhotoIDs); err != nil {
			return IncidentRow{}, false, err
		}
		report.IncidentID, report.Outcome = dup, domain.ReportJoined
		s.saveReport(ctx, report)
		row, err = s.GetIncident(ctx, dup, reporterID)
		return row, false, err
	}
	id, err := s.repo.CreateIncident(ctx, repository.NewIncident{
		HouseID: houseID, ReporterID: reporterID, Title: title, Description: d, Severity: sev,
		Entrance: in.Entrance, Riser: in.FloorZone, Category: in.Category, Routing: routing, SLA: domain.SLA(sev),
		PhotoIDs: in.PhotoIDs, Suspicious: falseAlarms >= domain.SuspiciousAfterFalseAlarms,
	})
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

func (s *Service) SetIncidentStatus(ctx context.Context, u model.AppUser, incidentID int64, to domain.IncidentStatus) (IncidentRow, error) {
	if err := s.CanAccessIncident(ctx, u, incidentID); err != nil {
		return IncidentRow{}, err
	}
	from, err := s.repo.FindIncidentStatus(ctx, incidentID)
	if err != nil {
		return IncidentRow{}, err
	}
	if !domain.CanDispatcherMove(from, to) {
		return IncidentRow{}, ErrInvalidStatus
	}
	moved, err := s.repo.TransitionIncident(ctx, incidentID, from, to, "incident_"+string(to), s.statusNotice(ukStatusText[to]))
	if err != nil {
		return IncidentRow{}, err
	}
	if !moved {
		return IncidentRow{}, ErrInvalidStatus
	}
	if r, err := s.repo.GetIncident(ctx, incidentID, u.ID); err == nil && r.ExternalID != nil {
		if err := s.uk.SetStatus(ctx, *r.ExternalID, to); err != nil {
			slog.Warn("uk: dispatcher status not synced", "incident", incidentID, "status", to, "err", err)
		}
	}
	return s.GetIncident(ctx, incidentID, u.ID)
}
