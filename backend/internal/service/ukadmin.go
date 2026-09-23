package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
	"ukapp/internal/ukauth"

	"ukapp/gen/db/ukapp/public/model"
)

type UkSession struct {
	Token     string
	ExpiresAt time.Time
	User      model.AppUser
	Org       model.Uk
}

func (s *Service) UseTokenKey(pemKey string) error {
	t, err := ukauth.NewTokenSigner(pemKey)
	if err != nil {
		return err
	}
	s.tokens = t
	return nil
}

func (s *Service) EsiaMockLogin(ctx context.Context, inn, password string, requesterMaxID *int64) (UkSession, error) {
	if !domain.ValidOrgINN(inn) || password == "" {
		return UkSession{}, ErrInvalidInput
	}
	failures, err := s.repo.LoginFailures(ctx, inn, time.Now().Add(-domain.UkLoginWindow))
	if err != nil {
		return UkSession{}, err
	}
	if failures >= domain.UkLoginMaxFailures {
		return UkSession{}, ErrRateLimited
	}
	org, user, err := s.checkUkCredentials(ctx, inn, password)
	var userID *int64
	if err == nil {
		userID = &user.ID
	}
	if aerr := s.repo.AddLoginAttempt(ctx, inn, requesterMaxID, userID, err == nil); aerr != nil {
		return UkSession{}, aerr
	}
	if err != nil {
		return UkSession{}, err
	}
	token, exp, err := s.tokens.Issue(user.ID, org.ID, inn, time.Now())
	if err != nil {
		return UkSession{}, err
	}
	slog.Info("uk dispatcher signed in", "user", user.ID, "uk", org.ID, "max_user", requesterMaxID)
	return UkSession{Token: token, ExpiresAt: exp, User: user, Org: org}, nil
}

func (s *Service) checkUkCredentials(ctx context.Context, inn, password string) (model.Uk, model.AppUser, error) {
	org, err := s.repo.UkByINN(ctx, inn)
	if errors.Is(err, ErrNotFound) {
		ukauth.PasswordMatches(nil, password)
		return model.Uk{}, model.AppUser{}, ErrForbidden
	}
	if err != nil {
		return model.Uk{}, model.AppUser{}, err
	}
	staff, err := s.repo.UkDispatchers(ctx, org.ID)
	if err != nil {
		return model.Uk{}, model.AppUser{}, err
	}
	var match *model.AppUser
	for i, u := range staff {
		if u.PasswordHash != nil && ukauth.PasswordMatches(u.PasswordHash, password) {
			match = &staff[i]
			break
		}
	}
	if match == nil {
		if len(staff) == 0 {
			ukauth.PasswordMatches(nil, password)
		}
		return model.Uk{}, model.AppUser{}, ErrForbidden
	}
	if !match.AdsAuthority || !domain.LicenseActive(org.LicenseNumber, org.LicenseValidUntil, time.Now()) {
		return model.Uk{}, model.AppUser{}, ErrForbidden
	}
	return org, *match, nil
}

func (s *Service) AuthenticateUkToken(ctx context.Context, raw string) (model.AppUser, error) {
	c, err := s.tokens.Parse(raw)
	if err != nil {
		return model.AppUser{}, ErrForbidden
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return model.AppUser{}, ErrForbidden
	}
	u, err := s.repo.GetUser(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return model.AppUser{}, ErrForbidden
	}
	if err != nil {
		return model.AppUser{}, err
	}
	if u.Role != domain.RoleUkDispatcher || u.UkID == nil || *u.UkID != c.UkID {
		return model.AppUser{}, ErrForbidden
	}
	return u, nil
}

func (s *Service) UnbindHouse(ctx context.Context, userID int64) error {
	return s.repo.UnbindHouse(ctx, userID)
}

func (s *Service) CanAccessIncident(ctx context.Context, u model.AppUser, incidentID int64) error {
	if u.Role != domain.RoleUkDispatcher {
		return nil
	}
	ukID, err := s.repo.IncidentUkID(ctx, incidentID)
	if err != nil {
		return err
	}
	if u.UkID == nil || *u.UkID != ukID {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UkQueue(ctx context.Context, u model.AppUser, offset, limit int64) ([]repository.UkQueueRow, int64, error) {
	return s.repo.UkQueue(ctx, *u.UkID, offset, limit)
}

type EventInput struct {
	HouseID                    int64
	Reason, Responsible        string
	Entrance, Riser            *string
	ScheduledFrom, ScheduledTo time.Time
}

func (s *Service) CreateEvent(ctx context.Context, u model.AppUser, in EventInput) (model.UkEvent, error) {
	if in.Reason == "" || in.Responsible == "" || !in.ScheduledTo.After(in.ScheduledFrom) {
		return model.UkEvent{}, ErrInvalidInput
	}
	ok, err := s.CanAccessHouse(ctx, u, in.HouseID)
	if err != nil {
		return model.UkEvent{}, err
	}
	if !ok {
		return model.UkEvent{}, ErrNotFound
	}
	return s.repo.InsertEvent(ctx, model.UkEvent{
		HouseID: in.HouseID, AuthorID: u.ID, Reason: in.Reason, Responsible: in.Responsible,
		Entrance: in.Entrance, Riser: in.Riser, ScheduledFrom: in.ScheduledFrom, ScheduledTo: in.ScheduledTo,
	})
}

func (s *Service) ListEvents(ctx context.Context, u model.AppUser, houseID int64) ([]model.UkEvent, error) {
	ok, err := s.CanAccessHouse(ctx, u, houseID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	return s.repo.ListHouseEvents(ctx, houseID, time.Now())
}
