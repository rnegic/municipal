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
	var avatar *string
	if iu.PhotoURL != "" {
		avatar = &iu.PhotoURL
	}
	var u model.AppUser
	err := AppUser.INSERT(AppUser.MaxUserID, AppUser.FullName, AppUser.AvatarURL).
		VALUES(iu.ID, name, avatar).
		ON_CONFLICT(AppUser.MaxUserID).DO_UPDATE(SET(
		AppUser.FullName.SET(AppUser.EXCLUDED.FullName),
		AppUser.AvatarURL.SET(AppUser.EXCLUDED.AvatarURL),
	)).
		RETURNING(AppUser.AllColumns).
		QueryContext(ctx, s.db, &u)
	return u, err
}

type UkContacts struct {
	Phone, EmergencyPhone, Email, Website, OfficeAddress, WorkingHours *string
}

func (s *Store) UpsertUk(ctx context.Context, externalID, name string, contacts UkContacts) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO uk (external_id, name, phone, emergency_phone, email, website, office_address, working_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (external_id) DO UPDATE SET
			name = EXCLUDED.name,
			phone = COALESCE(EXCLUDED.phone, uk.phone),
			emergency_phone = COALESCE(EXCLUDED.emergency_phone, uk.emergency_phone),
			email = COALESCE(EXCLUDED.email, uk.email),
			website = COALESCE(EXCLUDED.website, uk.website),
			office_address = COALESCE(EXCLUDED.office_address, uk.office_address),
			working_hours = COALESCE(EXCLUDED.working_hours, uk.working_hours)
		RETURNING id`,
		externalID, name, contacts.Phone, contacts.EmergencyPhone, contacts.Email,
		contacts.Website, contacts.OfficeAddress, contacts.WorkingHours).Scan(&id)
	return id, err
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

func (s *Store) SetUserHouse(ctx context.Context, userID, houseID int64) (bool, error) {
	res, err := AppUser.UPDATE(AppUser.HouseID).SET(Int64(houseID)).
		WHERE(AppUser.ID.EQ(Int64(userID)).AND(AppUser.HouseID.IS_NULL()).
			AND(EXISTS(SELECT(House.ID).FROM(House).WHERE(House.ID.EQ(Int64(houseID)))))).
		ExecContext(ctx, s.db)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (s *Store) FindHouse(ctx context.Context, id int64) (model.House, error) {
	var h model.House
	err := SELECT(House.ID, House.AddressRaw, House.UkID).FROM(House).WHERE(House.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &h)
	if errors.Is(err, qrm.ErrNoRows) {
		return model.House{}, ErrNotFound
	}
	return h, err
}

func (s *Store) HouseUk(ctx context.Context, houseID int64) (model.Uk, error) {
	var u model.Uk
	err := SELECT(Uk.AllColumns).
		FROM(Uk.INNER_JOIN(House, House.UkID.EQ(Uk.ID))).
		WHERE(House.ID.EQ(Int64(houseID))).
		QueryContext(ctx, s.db, &u)
	return u, notFound(err)
}
