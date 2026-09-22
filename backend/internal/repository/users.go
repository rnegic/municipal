package repository

import (
	"context"
	"errors"
	"strings"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

func (s *Store) UpsertUser(ctx context.Context, iu domain.InitUser) (model.AppUser, error) {
	name := strings.TrimSpace(iu.FirstName + " " + iu.LastName)
	var u model.AppUser
	err := AppUser.INSERT(AppUser.MaxUserID, AppUser.FullName).
		VALUES(iu.ID, name).
		ON_CONFLICT(AppUser.MaxUserID).DO_UPDATE(SET(AppUser.FullName.SET(AppUser.EXCLUDED.FullName))).
		RETURNING(AppUser.AllColumns).
		QueryContext(ctx, s.db, &u)
	return u, err
}

func (s *Store) UpsertUk(ctx context.Context, externalID, name string) (int64, error) {
	var u model.Uk
	err := Uk.INSERT(Uk.ExternalID, Uk.Name).
		VALUES(externalID, name).
		ON_CONFLICT(Uk.ExternalID).DO_UPDATE(SET(Uk.Name.SET(Uk.EXCLUDED.Name))).
		RETURNING(Uk.ID).
		QueryContext(ctx, s.db, &u)
	return u.ID, err
}

func (s *Store) BindHouse(ctx context.Context, userID int64, addressRaw, houseFiasID string, ukID int64, houseExternalID string) (int64, error) {
	var h model.House
	err := House.INSERT(House.AddressRaw, House.HouseFiasID, House.UkID, House.ExternalID).
		VALUES(addressRaw, houseFiasID, ukID, houseExternalID).
		ON_CONFLICT(House.HouseFiasID).DO_UPDATE(SET(
		House.AddressRaw.SET(House.EXCLUDED.AddressRaw),
		House.UkID.SET(House.EXCLUDED.UkID),
		House.ExternalID.SET(House.EXCLUDED.ExternalID),
	)).
		RETURNING(House.ID).
		QueryContext(ctx, s.db, &h)
	if err != nil {
		return 0, err
	}
	_, err = AppUser.UPDATE(AppUser.HouseID).SET(h.ID).WHERE(AppUser.ID.EQ(Int64(userID))).ExecContext(ctx, s.db)
	return h.ID, err
}

func (s *Store) FindHouse(ctx context.Context, id int64) (model.House, error) {
	var h model.House
	err := SELECT(House.ID, House.AddressRaw, House.UkID).FROM(House).WHERE(House.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &h)
	if errors.Is(err, qrm.ErrNoRows) {
		return model.House{}, ErrNotFound
	}
	return h, err
}
