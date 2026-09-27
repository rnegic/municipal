package repository

import (
	"context"
	"errors"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

type UnregisteredRow struct {
	model.Incident
	HouseExternalID *string
}

func (s *Store) UnregisteredIncidents(ctx context.Context, limit int64) ([]UnregisteredRow, error) {
	var rows []UnregisteredRow
	err := SELECT(Incident.AllColumns, House.ExternalID.AS("unregistered_row.house_external_id")).
		FROM(Incident.INNER_JOIN(House, House.ID.EQ(Incident.HouseID))).
		WHERE(Incident.ExternalID.IS_NULL().AND(Incident.MergedIntoID.IS_NULL()).AND(NOT(EXISTS(
			SELECT(UkAPIKey.ID).FROM(UkAPIKey).
				WHERE(UkAPIKey.UkID.EQ(House.UkID).AND(UkAPIKey.RevokedAt.IS_NULL())),
		)))).
		ORDER_BY(Incident.ID).LIMIT(limit).
		QueryContext(ctx, s.db, &rows)
	return rows, err
}

func (s *Store) SetIncidentExternalID(ctx context.Context, id int64, externalID string) error {
	_, err := Incident.UPDATE(Incident.ExternalID).SET(externalID).
		WHERE(Incident.ID.EQ(Int64(id))).ExecContext(ctx, s.db)
	return err
}

func (s *Store) ApplyUkStatus(ctx context.Context, externalID string, status domain.IncidentStatus, kind string, notify func(model.Incident) OutboxPayload) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	var inc model.Incident
	upd := Incident.UPDATE(Incident.Status).SET(string(status))
	if status.Closed() {
		upd = Incident.UPDATE(Incident.Status, Incident.ResolvedAt).SET(string(status), NOW())
	}
	err = upd.WHERE(
		Incident.ExternalID.EQ(String(externalID)).
			AND(Incident.Status.NOT_EQ(String(string(status)))).
			AND(Incident.Status.NOT_IN(String(string(domain.IncidentDone)), String(string(domain.IncidentFalseAlarm)))),
	).RETURNING(Incident.ID, Incident.Title).QueryContext(ctx, tx, &inc)
	if errors.Is(err, qrm.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	targets, err := s.SubscriberMaxIDs(ctx, inc.ID)
	if err != nil {
		return false, err
	}
	if err := s.Enqueue(ctx, tx, targets, kind, notify(inc)); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
