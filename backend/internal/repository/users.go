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

// UpsertUser creates the user on first visit; on later visits refreshes the name from MAX.
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

// BindHouse upserts the house by FIAS id (two residents of one building share a row)
// and attaches it to the user. uk_id = first UK in the table (demo: single UK).
func (s *Store) BindHouse(ctx context.Context, userID int64, addressRaw, houseFiasID string) (int64, error) {
	var h model.House
	err := House.INSERT(House.AddressRaw, House.HouseFiasID, House.UkID).
		VALUES(addressRaw, houseFiasID, SELECT(Uk.ID).FROM(Uk).ORDER_BY(Uk.ID).LIMIT(1)).
		ON_CONFLICT(House.HouseFiasID).DO_UPDATE(SET(House.AddressRaw.SET(House.AddressRaw))).
		RETURNING(House.ID).
		QueryContext(ctx, s.db, &h)
	if err != nil {
		return 0, err
	}
	_, err = AppUser.UPDATE(AppUser.HouseID).SET(h.ID).WHERE(AppUser.ID.EQ(Int64(userID))).ExecContext(ctx, s.db)
	return h.ID, err
}

// FindHouse returns ErrNotFound for a missing house.
func (s *Store) FindHouse(ctx context.Context, id int64) (model.House, error) {
	var h model.House
	err := SELECT(House.ID, House.AddressRaw).FROM(House).WHERE(House.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &h)
	if errors.Is(err, qrm.ErrNoRows) {
		return model.House{}, ErrNotFound
	}
	return h, err
}
