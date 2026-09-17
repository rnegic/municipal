package repository

import (
	"context"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

// FindIncidentStatus returns ErrNotFound for a missing incident.
func (s *Store) FindIncidentStatus(ctx context.Context, incidentID int64) (domain.IncidentStatus, error) {
	var inc model.Incident
	err := SELECT(Incident.Status).FROM(Incident).WHERE(Incident.ID.EQ(Int64(incidentID))).QueryContext(ctx, s.db, &inc)
	if errors.Is(err, qrm.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return domain.IncidentStatus(inc.Status), nil
}

// UpsertConfirmation is idempotent; returns the (first or existing) confirmed_at.
func (s *Store) UpsertConfirmation(ctx context.Context, incidentID, userID int64) (time.Time, error) {
	var row model.IncidentConfirmation
	err := IncidentConfirmation.INSERT(IncidentConfirmation.IncidentID, IncidentConfirmation.UserID).
		VALUES(incidentID, userID).
		ON_CONFLICT(IncidentConfirmation.IncidentID, IncidentConfirmation.UserID).
		DO_UPDATE(SET(IncidentConfirmation.IncidentID.SET(IncidentConfirmation.EXCLUDED.IncidentID))).
		RETURNING(IncidentConfirmation.ConfirmedAt).
		QueryContext(ctx, s.db, &row)
	return row.ConfirmedAt, err
}

func (s *Store) ConfirmationCount(ctx context.Context, incidentID int64) (int, error) {
	var cnt struct{ Count int }
	err := SELECT(COUNT(IncidentConfirmation.UserID).AS("count")).FROM(IncidentConfirmation).
		WHERE(IncidentConfirmation.IncidentID.EQ(Int64(incidentID))).QueryContext(ctx, s.db, &cnt)
	return cnt.Count, err
}

// CloseIfVerifying atomically moves the incident from verifying to done and enqueues the
// notification in the same transaction; closed=false if the incident wasn't in verifying
// any more (already closed by a racing confirmation).
func (s *Store) CloseIfVerifying(ctx context.Context, incidentID int64, kind string, payload OutboxPayload) (closed bool, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	res, err := Incident.UPDATE(Incident.Status, Incident.ResolvedAt).SET(string(domain.IncidentDone), NOW()).
		WHERE(Incident.ID.EQ(Int64(incidentID)).AND(Incident.Status.EQ(String(string(domain.IncidentVerifying))))).
		ExecContext(ctx, tx)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil
	}
	targets, err := s.SubscriberMaxIDs(ctx, incidentID)
	if err != nil {
		return false, err
	}
	if err := s.Enqueue(ctx, tx, targets, kind, payload); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
