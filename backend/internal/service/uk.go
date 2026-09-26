package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"

	"ukapp/gen/db/ukapp/public/model"
)

type UkProvider interface {
	FindHouse(ctx context.Context, fiasID string) (UkHouse, error)
	RegisterIncident(ctx context.Context, in UkIncident) (id string, status domain.IncidentStatus, err error)
	IncidentUpdates(ctx context.Context, since time.Time) ([]UkIncidentUpdate, error)
	SetStatus(ctx context.Context, id string, status domain.IncidentStatus) error
}

var ErrUkHouseNotFound = errors.New("house not served by uk")

// UkOrg — организация из системы УК; nil-контакты = система УК их не передала.
type UkOrg struct {
	ExternalID, Name                      string
	Phone, EmergencyPhone, Email, Website *string
	OfficeAddress, WorkingHours           *string
}

type UkHouse struct {
	ID, Address string
	Org         UkOrg
}

type UkIncident struct {
	ExternalRef, HouseID, Title, Description, Severity string
	Entrance, Riser                                    *string
	Suspicious                                         bool
}

type UkIncidentUpdate struct {
	ID, ExternalRef string
	Status          domain.IncidentStatus
	UpdatedAt       time.Time
}

func (s *Service) CanAccessHouse(ctx context.Context, u model.AppUser, houseID int64) (bool, error) {
	if u.Role != domain.RoleUkDispatcher {
		return u.HouseID != nil && *u.HouseID == houseID, nil
	}
	h, err := s.repo.FindHouse(ctx, houseID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil && u.UkID != nil && *u.UkID == h.UkID, err
}

func (s *Service) HouseStats(ctx context.Context, houseID int64) (repository.HouseStats, error) {
	return s.repo.HouseStats(ctx, houseID)
}

const (
	MaxPhotosPerIncident = 5
	MaxPhotosPerUserHour = 20
)

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
	return s.repo.InsertPhoto(ctx, &incidentID, userID, contentType, data)
}

func (s *Service) PhotoRateLimited(ctx context.Context, userID int64) (bool, error) {
	_, byUser, err := s.repo.PhotoCounts(ctx, 0, userID, time.Hour)
	return byUser >= MaxPhotosPerUserHour, err
}

func (s *Service) AddStagedPhoto(ctx context.Context, userID int64, contentType string, data []byte) (int64, error) {
	_, byUser, err := s.repo.PhotoCounts(ctx, 0, userID, time.Hour)
	if err != nil {
		return 0, err
	}
	if byUser >= MaxPhotosPerUserHour {
		return 0, ErrRateLimited
	}
	return s.repo.InsertPhoto(ctx, nil, userID, contentType, data)
}

func (s *Service) RunPhotoCleanup(ctx context.Context) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		if n, err := s.repo.DeleteStalePhotos(ctx, 24*time.Hour); err != nil {
			slog.Warn("stale photo cleanup failed", "err", err)
		} else if n > 0 {
			slog.Info("stale photos deleted", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *Service) PhotoIDs(ctx context.Context, incidentIDs []int64) (map[int64][]int64, error) {
	return s.repo.PhotoIDsByIncident(ctx, incidentIDs)
}

func (s *Service) GetPhoto(ctx context.Context, id int64) (model.IncidentPhoto, error) {
	return s.repo.GetPhoto(ctx, id)
}
