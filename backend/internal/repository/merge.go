package repository

import (
	"context"
	"database/sql"

	"ukapp/internal/domain"
)

type MergeResult struct {
	Merged        []int64
	ExternalIDs   []string
	AffectedCount int
}

func (s *Store) MergeIncidents(ctx context.Context, ukID, target int64, sources []int64, notify OutboxPayload) (MergeResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MergeResult{}, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	found, err := lockMergeCandidates(ctx, tx, append([]int64{target}, sources...))
	if err != nil {
		return MergeResult{}, err
	}
	toMerge, err := domain.PlanMerge(ukID, target, sources, found)
	if err != nil {
		return MergeResult{}, err
	}

	res := MergeResult{Merged: toMerge}
	if len(toMerge) > 0 {
		targets, err := queryCol[int64](ctx, tx, `
			SELECT DISTINCT u.max_user_id FROM incident_subscription s JOIN app_user u ON u.id = s.user_id
			WHERE s.incident_id = ANY($1) AND u.max_user_id IS NOT NULL`, toMerge)
		if err != nil {
			return MergeResult{}, err
		}

		for _, q := range []string{
			`INSERT INTO incident_subscription (incident_id, user_id, joined_at)
			 SELECT $1, user_id, min(joined_at) FROM incident_subscription WHERE incident_id = ANY($2) GROUP BY user_id
			 ON CONFLICT DO NOTHING`,
			`INSERT INTO incident_confirmation (incident_id, user_id, confirmed_at)
			 SELECT $1, user_id, min(confirmed_at) FROM incident_confirmation WHERE incident_id = ANY($2) GROUP BY user_id
			 ON CONFLICT DO NOTHING`,
			`UPDATE incident_photo SET incident_id = $1 WHERE incident_id = ANY($2)`,
			`INSERT INTO incident_report (incident_id, reporter_id, house_id, title, description, severity, entrance, riser, outcome, dedup_version)
			 SELECT $1, reporter_id, house_id, title, description, severity, entrance, riser, '` + domain.ReportMergedManual + `', '` + domain.MergeDedupVersion + `'
			 FROM incident WHERE id = ANY($2)
			 ON CONFLICT DO NOTHING`,
			`UPDATE incident SET merged_count = merged_count + (SELECT count(*) + coalesce(sum(merged_count), 0) FROM incident WHERE id = ANY($2))
			 WHERE id = $1`,
		} {
			if _, err := tx.ExecContext(ctx, q, target, toMerge); err != nil {
				return MergeResult{}, err
			}
		}
		if err := s.Enqueue(ctx, tx, targets, "incident_merged", notify); err != nil {
			return MergeResult{}, err
		}
		res.ExternalIDs, err = queryCol[string](ctx, tx, `
			UPDATE incident SET status = 'done', resolved_at = NOW(), merged_into_id = $1
			WHERE id = ANY($2) RETURNING coalesce(external_id, '')`, target, toMerge)
		if err != nil {
			return MergeResult{}, err
		}
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT count(DISTINCT user_id) FROM incident_subscription WHERE incident_id = $1`, target).Scan(&res.AffectedCount); err != nil {
		return MergeResult{}, err
	}
	return res, tx.Commit()
}

func queryCol[T any](ctx context.Context, tx *sql.Tx, q string, args ...any) ([]T, error) {
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []T
	for rows.Next() {
		var v T
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func lockMergeCandidates(ctx context.Context, tx *sql.Tx, ids []int64) ([]domain.MergeCandidate, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT i.id, i.house_id, h.uk_id, i.status, i.merged_into_id
		FROM incident i JOIN house h ON h.id = i.house_id
		WHERE i.id = ANY($1) ORDER BY i.id FOR UPDATE OF i`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MergeCandidate
	for rows.Next() {
		var c domain.MergeCandidate
		if err := rows.Scan(&c.ID, &c.HouseID, &c.UkID, &c.Status, &c.MergedIntoID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
