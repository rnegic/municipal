package service

import (
	"context"
	"errors"
	"time"

	"ukapp/internal/repository"

	"ukapp/gen/db/ukapp/public/model"
)

// CanAccessHouse: a resident sees their own house, a dispatcher — every house of their UK.
func (s *Service) CanAccessHouse(ctx context.Context, u model.AppUser, houseID int64) (bool, error) {
	if u.HouseID != nil && *u.HouseID == houseID {
		return true, nil
	}
	if u.Role != "uk_dispatcher" || u.UkID == nil {
		return false, nil
	}
	h, err := s.repo.FindHouse(ctx, houseID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return h.UkID == *u.UkID, err
}

func (s *Service) HouseStats(ctx context.Context, houseID int64) (repository.HouseStats, error) {
	return s.repo.HouseStats(ctx, houseID)
}

// UkQueue: empty for a dispatcher without a UK (nothing to serve).
func (s *Service) UkQueue(ctx context.Context, u model.AppUser, offset, limit int64) ([]repository.UkQueueRow, int64, error) {
	if u.UkID == nil {
		return nil, 0, nil
	}
	return s.repo.UkQueue(ctx, *u.UkID, offset, limit)
}

// CreateEvent: the house must belong to the dispatcher's UK (ErrNotFound otherwise).
func (s *Service) CreateEvent(ctx context.Context, u model.AppUser, houseID int64, entrance, riser *string, reason, responsible string, from, to time.Time) (model.Event, error) {
	if reason == "" || responsible == "" || !to.After(from) {
		return model.Event{}, ErrInvalidInput
	}
	ok, err := s.CanAccessHouse(ctx, u, houseID)
	if err != nil {
		return model.Event{}, err
	}
	if !ok {
		return model.Event{}, ErrNotFound
	}
	return s.repo.CreateEvent(ctx, model.Event{
		HouseID: houseID, UkDispatcherID: u.ID, Entrance: entrance, Riser: riser,
		Reason: reason, Responsible: responsible, ScheduledFrom: from, ScheduledTo: to,
	})
}

// GetEvent returns ErrNotFound both for a missing event and for one in a house the user can't see.
func (s *Service) GetEvent(ctx context.Context, u model.AppUser, id int64) (model.Event, error) {
	e, err := s.repo.GetEvent(ctx, id)
	if err != nil {
		return model.Event{}, err
	}
	ok, err := s.CanAccessHouse(ctx, u, e.HouseID)
	if err != nil {
		return model.Event{}, err
	}
	if !ok {
		return model.Event{}, ErrNotFound
	}
	return e, nil
}

func (s *Service) ListEvents(ctx context.Context, houseID int64, activeOnly bool) ([]model.Event, error) {
	return s.repo.ListEvents(ctx, houseID, activeOnly)
}

// UkHouses: empty for a dispatcher without a UK.
func (s *Service) UkHouses(ctx context.Context, u model.AppUser) ([]model.House, error) {
	if u.UkID == nil {
		return nil, nil
	}
	return s.repo.UkHouses(ctx, *u.UkID)
}

const (
	MaxPhotosPerIncident = 5
	MaxPhotosPerUserHour = 20
)

// AddPhoto stores already-validated (transport sniffs type and caps size) image bytes.
// Only the reporter or a subscriber may attach (ErrForbidden); caps: MaxPhotosPerIncident
// (ErrInvalidStatus) and MaxPhotosPerUserHour (ErrRateLimited).
// ponytail: count-based limits in Postgres, no token bucket; enough for a webview form.
func (s *Service) AddPhoto(ctx context.Context, incidentID, userID int64, contentType string, data []byte) (int64, error) {
	inc, err := s.repo.GetIncident(ctx, incidentID, userID)
	if err != nil {
		return 0, err
	}
	if inc.ReporterID != userID && !inc.JoinedByMe {
		return 0, ErrForbidden
	}
	byIncident, byUser, err := s.repo.PhotoCounts(ctx, incidentID, userID, time.Hour)
	if err != nil {
		return 0, err
	}
	if byUser >= MaxPhotosPerUserHour {
		return 0, ErrRateLimited
	}
	if byIncident >= MaxPhotosPerIncident {
		return 0, ErrInvalidStatus
	}
	return s.repo.InsertPhoto(ctx, incidentID, userID, contentType, data)
}

// PhotoIDs returns photo ids grouped by incident (for list views that aren't IncidentRow).
func (s *Service) PhotoIDs(ctx context.Context, incidentIDs []int64) (map[int64][]int64, error) {
	return s.repo.PhotoIDsByIncident(ctx, incidentIDs)
}

func (s *Service) GetPhoto(ctx context.Context, id int64) (model.IncidentPhoto, error) {
	return s.repo.GetPhoto(ctx, id)
}
