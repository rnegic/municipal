package repository

import (
	"context"
	"time"
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
