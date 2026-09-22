package service

import (
	"context"
	"errors"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
)

var (
	ErrAddressNotResolved = errors.New("address not resolved")
	ErrHouseNotServed     = errors.New("house not served by connected uk")
)

func (s *Service) UpsertUser(ctx context.Context, iu domain.InitUser) (model.AppUser, error) {
	return s.repo.UpsertUser(ctx, iu)
}

// BindHouse normalizes a free-text address via DaData, asks the UK system who serves the
// building, and binds the house to the user. Unknown to the UK → ErrHouseNotServed.
func (s *Service) BindHouse(ctx context.Context, userID int64, rawAddress string) (houseID int64, address string, err error) {
	sugs, err := s.dd.Suggest(ctx, rawAddress, DefaultSuggestCount)
	if err != nil {
		return 0, "", err
	}
	if len(sugs) == 0 {
		return 0, "", ErrAddressNotResolved
	}
	best := sugs[0]
	h, err := s.uk.FindHouse(ctx, best.HouseFiasID)
	if errors.Is(err, ErrUkHouseNotFound) {
		return 0, "", ErrHouseNotServed
	}
	if err != nil {
		return 0, "", err
	}
	ukID, err := s.repo.UpsertUk(ctx, h.OrgID, h.OrgName)
	if err != nil {
		return 0, "", err
	}
	houseID, err = s.repo.BindHouse(ctx, userID, best.Value, best.HouseFiasID, ukID, h.ID)
	return houseID, best.Value, err
}

func (s *Service) HouseAddress(ctx context.Context, houseID int64) (string, error) {
	h, err := s.repo.FindHouse(ctx, houseID)
	return h.AddressRaw, err
}
