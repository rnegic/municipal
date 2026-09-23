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

var openStatuses = []Expression{String(string(domain.IncidentAccepted)), String(string(domain.IncidentInProgress))}
var activeStatuses = []Expression{
	String(string(domain.IncidentAccepted)), String(string(domain.IncidentInProgress)), String(string(domain.IncidentVerifying)),
}

func (s *Store) OpenIncidents(ctx context.Context, houseID int64) ([]domain.OpenIncident, error) {
	var rows []model.Incident
	err := SELECT(Incident.ID, Incident.HouseID, Incident.Title, Incident.Riser, Incident.Status, Incident.CreatedAt).
		FROM(Incident).
		WHERE(Incident.HouseID.EQ(Int64(houseID)).AND(Incident.Status.IN(openStatuses...))).
		QueryContext(ctx, s.db, &rows)
	if err != nil {
		return nil, err
	}
	out := make([]domain.OpenIncident, len(rows))
	for i, r := range rows {
		out[i] = domain.OpenIncident{ID: r.ID, HouseID: r.HouseID, Title: r.Title, Riser: r.Riser, Status: domain.IncidentStatus(r.Status), CreatedAt: r.CreatedAt}
	}
	return out, nil
}

func (s *Store) CreateIncident(ctx context.Context, houseID, reporterID int64, title, description string, severity domain.Severity, entrance, riser *string, sla time.Duration) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit
	var inc model.Incident
	err = Incident.INSERT(Incident.HouseID, Incident.Title, Incident.Severity, Incident.ReporterID, Incident.Description, Incident.Entrance, Incident.Riser, Incident.DueAt).
		VALUES(houseID, title, string(severity), reporterID, description, entrance, riser, NOW().ADD(INTERVALd(sla))).
		RETURNING(Incident.ID).
		QueryContext(ctx, tx, &inc)
	if err != nil {
		return 0, err
	}
	_, err = IncidentSubscription.INSERT(IncidentSubscription.IncidentID, IncidentSubscription.UserID).
		VALUES(inc.ID, reporterID).ExecContext(ctx, tx)
	if err != nil {
		return 0, err
	}
	return inc.ID, tx.Commit()
}

func (s *Store) Subscribe(ctx context.Context, incidentID, userID int64) error {
	_, err := IncidentSubscription.INSERT(IncidentSubscription.IncidentID, IncidentSubscription.UserID).
		VALUES(incidentID, userID).
		ON_CONFLICT().DO_NOTHING().
		ExecContext(ctx, s.db)
	if err != nil && isFKViolation(err) {
		return ErrNotFound
	}
	return err
}

type IncidentRow struct {
	model.Incident
	Subscribers   int
	JoinedByMe    bool
	ConfirmedByMe bool
}

func incidentSelect(userID int64) SelectStatement {
	subs := IncidentSubscription.AS("subs")
	conf := IncidentConfirmation.AS("conf")
	return SELECT(
		Incident.AllColumns,
		SELECT(COUNT(IncidentSubscription.UserID)).FROM(IncidentSubscription).
			WHERE(IncidentSubscription.IncidentID.EQ(Incident.ID)).AS("incident_row.subscribers"),
		EXISTS(SELECT(subs.UserID).FROM(subs).
			WHERE(subs.IncidentID.EQ(Incident.ID).AND(subs.UserID.EQ(Int64(userID))))).AS("incident_row.joined_by_me"),
		EXISTS(SELECT(conf.UserID).FROM(conf).
			WHERE(conf.IncidentID.EQ(Incident.ID).AND(conf.UserID.EQ(Int64(userID))))).AS("incident_row.confirmed_by_me"),
	).FROM(Incident)
}

func (s *Store) GetIncident(ctx context.Context, id, userID int64) (IncidentRow, error) {
	var row IncidentRow
	err := incidentSelect(userID).WHERE(Incident.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return IncidentRow{}, ErrNotFound
	}
	return row, err
}

func (s *Store) ListActiveIncidents(ctx context.Context, houseID, userID int64) ([]IncidentRow, error) {
	var rows []IncidentRow
	err := incidentSelect(userID).
		WHERE(Incident.HouseID.EQ(Int64(houseID)).AND(Incident.Status.IN(activeStatuses...))).
		ORDER_BY(Incident.Severity.EQ(String(string(domain.SeverityCritical))).DESC(), Incident.CreatedAt.DESC()).
		QueryContext(ctx, s.db, &rows)
	return rows, err
}

func (s *Store) ListUserIncidents(ctx context.Context, houseID, userID int64, offset, limit int64) ([]IncidentRow, int64, error) {
	mine := Incident.HouseID.EQ(Int64(houseID)).AND(
		Incident.ReporterID.EQ(Int64(userID)).OR(EXISTS(
			SELECT(IncidentSubscription.UserID).FROM(IncidentSubscription).
				WHERE(IncidentSubscription.IncidentID.EQ(Incident.ID).AND(IncidentSubscription.UserID.EQ(Int64(userID)))))))
	var total struct{ Count int64 }
	err := SELECT(COUNT(Incident.ID).AS("count")).FROM(Incident).WHERE(mine).QueryContext(ctx, s.db, &total)
	if err != nil {
		return nil, 0, err
	}
	var rows []IncidentRow
	err = incidentSelect(userID).
		WHERE(mine).
		ORDER_BY(Incident.CreatedAt.DESC()).
		OFFSET(offset).LIMIT(limit).
		QueryContext(ctx, s.db, &rows)
	return rows, total.Count, err
}

func (s *Store) SubscriberMaxIDs(ctx context.Context, incidentID int64) ([]int64, error) {
	var users []model.AppUser
	err := SELECT(AppUser.MaxUserID).
		FROM(IncidentSubscription.INNER_JOIN(AppUser, AppUser.ID.EQ(IncidentSubscription.UserID))).
		WHERE(IncidentSubscription.IncidentID.EQ(Int64(incidentID))).
		QueryContext(ctx, s.db, &users)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		return nil, err
	}
	out := make([]int64, 0, len(users))
	for _, u := range users {
		if u.MaxUserID != nil {
			out = append(out, *u.MaxUserID)
		}
	}
	return out, nil
}

func (s *Store) SubscriberCount(ctx context.Context, incidentID int64) (int, error) {
	var cnt struct{ Count int }
	err := SELECT(COUNT(IncidentSubscription.UserID).AS("count")).FROM(IncidentSubscription).
		WHERE(IncidentSubscription.IncidentID.EQ(Int64(incidentID))).QueryContext(ctx, s.db, &cnt)
	return cnt.Count, err
}
