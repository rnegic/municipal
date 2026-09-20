package repository

import (
	"context"
	"time"

	. "github.com/go-jet/jet/v2/postgres"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

type HouseStats struct {
	ActiveIncidents    int
	InProgress         int
	ResolvedLast30Days int
	AvgResolutionHours *float64
	LastIncidentAt     *time.Time
}

// HouseStats aggregates the house's incidents in one scan.
func (s *Store) HouseStats(ctx context.Context, houseID int64) (HouseStats, error) {
	var st HouseStats
	err := s.db.QueryRowContext(ctx, `
		SELECT count(*) FILTER (WHERE status IN ('accepted','in_progress','verifying')),
		       count(*) FILTER (WHERE status = 'in_progress'),
		       count(*) FILTER (WHERE status = 'done' AND resolved_at >= now() - interval '30 days'),
		       avg(EXTRACT(EPOCH FROM resolved_at - created_at) / 3600)::float8,
		       max(created_at)
		FROM incident WHERE house_id = $1`, houseID).
		Scan(&st.ActiveIncidents, &st.InProgress, &st.ResolvedLast30Days, &st.AvgResolutionHours, &st.LastIncidentAt)
	return st, err
}

// UkHouses lists the houses served by the UK, by id.
func (s *Store) UkHouses(ctx context.Context, ukID int64) ([]model.House, error) {
	var out []model.House
	err := SELECT(House.ID, House.AddressRaw).FROM(House).WHERE(House.UkID.EQ(Int64(ukID))).ORDER_BY(House.ID).
		QueryContext(ctx, s.db, &out)
	return out, err
}

// UkQueueRow = incident + house address, reporter name and counters; aliases must match field names.
type UkQueueRow struct {
	model.Incident
	HouseAddress  string
	ReporterName  string
	Subscribers   int
	Confirmations int
}

// UkQueue lists not-yet-done incidents of all houses served by the UK, critical first,
// newest first, with the total for pagination.
func (s *Store) UkQueue(ctx context.Context, ukID, offset, limit int64) ([]UkQueueRow, int64, error) {
	from := Incident.INNER_JOIN(House, House.ID.EQ(Incident.HouseID)).INNER_JOIN(AppUser, AppUser.ID.EQ(Incident.ReporterID))
	where := House.UkID.EQ(Int64(ukID)).AND(Incident.Status.NOT_EQ(String(string(domain.IncidentDone))))

	var total struct{ Count int64 }
	if err := SELECT(COUNT(Incident.ID).AS("count")).FROM(from).WHERE(where).QueryContext(ctx, s.db, &total); err != nil {
		return nil, 0, err
	}
	var rows []UkQueueRow
	err := SELECT(
		Incident.AllColumns,
		House.AddressRaw.AS("uk_queue_row.house_address"),
		AppUser.FullName.AS("uk_queue_row.reporter_name"),
		SELECT(COUNT(IncidentSubscription.UserID)).FROM(IncidentSubscription).
			WHERE(IncidentSubscription.IncidentID.EQ(Incident.ID)).AS("uk_queue_row.subscribers"),
		SELECT(COUNT(IncidentConfirmation.UserID)).FROM(IncidentConfirmation).
			WHERE(IncidentConfirmation.IncidentID.EQ(Incident.ID)).AS("uk_queue_row.confirmations"),
	).FROM(from).WHERE(where).
		ORDER_BY(Incident.Severity.EQ(String(string(domain.SeverityCritical))).DESC(), Incident.CreatedAt.DESC()).
		OFFSET(offset).LIMIT(limit).
		QueryContext(ctx, s.db, &rows)
	return rows, total.Count, err
}
