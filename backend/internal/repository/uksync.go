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

// UnregisteredRow = incident + house.external_id; alias must match the field name.
type UnregisteredRow struct {
	model.Incident
	HouseExternalID *string
}

// UnregisteredIncidents returns incidents not yet known to the UK system, oldest first.
func (s *Store) UnregisteredIncidents(ctx context.Context, limit int64) ([]UnregisteredRow, error) {
	var rows []UnregisteredRow
	err := SELECT(Incident.AllColumns, House.ExternalID.AS("unregistered_row.house_external_id")).
		FROM(Incident.INNER_JOIN(House, House.ID.EQ(Incident.HouseID))).
		WHERE(Incident.ExternalID.IS_NULL()).
		ORDER_BY(Incident.ID).LIMIT(limit).
		QueryContext(ctx, s.db, &rows)
	return rows, err
}

func (s *Store) SetIncidentExternalID(ctx context.Context, id int64, externalID string) error {
	_, err := Incident.UPDATE(Incident.ExternalID).SET(externalID).
		WHERE(Incident.ID.EQ(Int64(id))).ExecContext(ctx, s.db)
	return err
}

// ApplyUkStatus applies a status reported by the UK system and enqueues the push to
// subscribers in the same transaction. changed=false when the incident is unknown, already
// in that status, or already done (residents' verdict is final).
func (s *Store) ApplyUkStatus(ctx context.Context, externalID string, status domain.IncidentStatus, kind string, payload OutboxPayload) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	var inc model.Incident
	upd := Incident.UPDATE(Incident.Status).SET(string(status))
	if status == domain.IncidentDone {
		upd = Incident.UPDATE(Incident.Status, Incident.ResolvedAt).SET(string(status), NOW())
	}
	err = upd.WHERE(
		Incident.ExternalID.EQ(String(externalID)).
			AND(Incident.Status.NOT_EQ(String(string(status)))).
			AND(Incident.Status.NOT_EQ(String(string(domain.IncidentDone)))),
	).RETURNING(Incident.ID).QueryContext(ctx, tx, &inc)
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
	if err := s.Enqueue(ctx, tx, targets, kind, payload); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
