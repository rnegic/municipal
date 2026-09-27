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

func (s *Store) TransitionIncident(ctx context.Context, incidentID int64, from, to domain.IncidentStatus, kind string, notify func(model.Incident) OutboxPayload) (moved bool, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	upd := Incident.UPDATE(Incident.Status).SET(string(to))
	if to == domain.IncidentDone {
		upd = Incident.UPDATE(Incident.Status, Incident.ResolvedAt).SET(string(to), NOW())
	}
	var inc model.Incident
	err = upd.WHERE(Incident.ID.EQ(Int64(incidentID)).AND(Incident.Status.EQ(String(string(from))))).
		RETURNING(Incident.ID, Incident.Title).QueryContext(ctx, tx, &inc)
	if errors.Is(err, qrm.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if kind != "" {
		targets, err := s.SubscriberMaxIDs(ctx, incidentID)
		if err != nil {
			return false, err
		}
		if err := s.Enqueue(ctx, tx, targets, kind, notify(inc)); err != nil {
			return false, err
		}
	}
	return true, tx.Commit()
}
