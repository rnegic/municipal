package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

func notFound(err error) error {
	if errors.Is(err, qrm.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *Store) GetUser(ctx context.Context, id int64) (model.AppUser, error) {
	var u model.AppUser
	err := SELECT(AppUser.AllColumns).FROM(AppUser).WHERE(AppUser.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &u)
	return u, notFound(err)
}

func (s *Store) UkByINN(ctx context.Context, inn string) (model.Uk, error) {
	var u model.Uk
	err := SELECT(Uk.AllColumns).FROM(Uk).WHERE(Uk.Inn.EQ(String(inn))).QueryContext(ctx, s.db, &u)
	return u, notFound(err)
}

func (s *Store) UkDispatchers(ctx context.Context, ukID int64) ([]model.AppUser, error) {
	var us []model.AppUser
	err := SELECT(AppUser.AllColumns).FROM(AppUser).
		WHERE(AppUser.UkID.EQ(Int64(ukID)).AND(AppUser.Role.EQ(String(domain.RoleUkDispatcher)))).
		QueryContext(ctx, s.db, &us)
	return us, err
}

func (s *Store) LoginFailures(ctx context.Context, inn string, since time.Time) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM uk_login_attempt WHERE inn = $1 AND NOT success AND created_at >= $2`, inn, since).Scan(&n)
	return n, err
}

func (s *Store) AddLoginAttempt(ctx context.Context, inn string, maxUserID, userID *int64, success bool) error {
	_, err := UkLoginAttempt.INSERT(UkLoginAttempt.Inn, UkLoginAttempt.MaxUserID, UkLoginAttempt.UserID, UkLoginAttempt.Success).
		VALUES(inn, maxUserID, userID, success).ExecContext(ctx, s.db)
	return err
}

func (s *Store) UnbindHouse(ctx context.Context, userID int64) error {
	_, err := AppUser.UPDATE(AppUser.HouseID).SET(NULL).WHERE(AppUser.ID.EQ(Int64(userID))).ExecContext(ctx, s.db)
	return err
}

func (s *Store) IncidentUkID(ctx context.Context, incidentID int64) (int64, error) {
	var ukID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT h.uk_id FROM incident i JOIN house h ON h.id = i.house_id WHERE i.id = $1`, incidentID).Scan(&ukID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return ukID, err
}

type UkQueueRow struct {
	ID             int64
	HouseID        int64
	HouseAddress   string
	Title          string
	Description    string
	Severity       string
	Status         string
	CreatedAt      time.Time
	DueAt          *time.Time
	AffectedCount  int
	ConfirmedCount int
	ReporterName   string
}

func (s *Store) UkQueue(ctx context.Context, ukID, offset, limit int64) ([]UkQueueRow, int64, error) {
	var total int64
	err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM incident i JOIN house h ON h.id = i.house_id WHERE h.uk_id = $1`, ukID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT i.id, i.house_id, h.address_raw, i.title, i.description, i.severity, i.status, i.created_at, i.due_at,
		       (SELECT count(*) FROM incident_subscription s WHERE s.incident_id = i.id),
		       (SELECT count(*) FROM incident_confirmation c WHERE c.incident_id = i.id),
		       u.full_name
		FROM incident i
		JOIN house h ON h.id = i.house_id
		JOIN app_user u ON u.id = i.reporter_id
		WHERE h.uk_id = $1
		ORDER BY i.created_at DESC, i.id DESC
		OFFSET $2 LIMIT $3`, ukID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []UkQueueRow
	for rows.Next() {
		var r UkQueueRow
		if err := rows.Scan(&r.ID, &r.HouseID, &r.HouseAddress, &r.Title, &r.Description, &r.Severity, &r.Status,
			&r.CreatedAt, &r.DueAt, &r.AffectedCount, &r.ConfirmedCount, &r.ReporterName); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (s *Store) InsertEvent(ctx context.Context, e model.UkEvent) (model.UkEvent, error) {
	var out model.UkEvent
	err := UkEvent.INSERT(UkEvent.HouseID, UkEvent.AuthorID, UkEvent.Reason, UkEvent.Responsible,
		UkEvent.Entrance, UkEvent.Riser, UkEvent.ScheduledFrom, UkEvent.ScheduledTo).
		MODEL(e).RETURNING(UkEvent.AllColumns).QueryContext(ctx, s.db, &out)
	return out, err
}

func (s *Store) ListHouseEvents(ctx context.Context, houseID int64, now time.Time) ([]model.UkEvent, error) {
	var out []model.UkEvent
	err := SELECT(UkEvent.AllColumns).FROM(UkEvent).
		WHERE(UkEvent.HouseID.EQ(Int64(houseID)).AND(UkEvent.ScheduledTo.GT(TimestampzT(now)))).
		ORDER_BY(UkEvent.ScheduledFrom.ASC(), UkEvent.ID.ASC()).
		QueryContext(ctx, s.db, &out)
	return out, err
}
