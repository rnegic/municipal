package service

import (
	"context"
	"errors"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
)

var ErrAddressNotResolved = errors.New("address not resolved")

func (s *Service) UpsertUser(ctx context.Context, iu domain.InitUser) (model.AppUser, error) {
	return s.repo.UpsertUser(ctx, iu)
}

// BindHouse normalizes a free-text address via DaData and binds the first (best) match.
func (s *Service) BindHouse(ctx context.Context, userID int64, rawAddress string) (houseID int64, address string, err error) {
	sugs, err := s.dd.Suggest(ctx, rawAddress)
	if err != nil {
		return 0, "", err
	}
	if len(sugs) == 0 {
		return 0, "", ErrAddressNotResolved
	}
	best := sugs[0]
	houseID, err = s.repo.BindHouse(ctx, userID, best.Value, best.HouseFiasID)
	return houseID, best.Value, err
}

func (s *Service) HouseAddress(ctx context.Context, houseID int64) (string, error) {
	h, err := s.repo.FindHouse(ctx, houseID)
	return h.AddressRaw, err
}
