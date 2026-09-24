package repository

import (
	"context"
	"errors"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

func (s *Store) InsertPhoto(ctx context.Context, incidentID *int64, userID int64, contentType string, data []byte) (int64, error) {
	var p model.IncidentPhoto
	err := IncidentPhoto.INSERT(IncidentPhoto.IncidentID, IncidentPhoto.UserID, IncidentPhoto.ContentType, IncidentPhoto.Data).
		VALUES(incidentID, userID, contentType, data).
		RETURNING(IncidentPhoto.ID).
		QueryContext(ctx, s.db, &p)
	if err != nil && isFKViolation(err) {
		return 0, ErrNotFound
	}
	return p.ID, err
}

func (s *Store) PhotoCounts(ctx context.Context, incidentID, userID int64, window time.Duration) (byIncident, byUserRecent int, err error) {
	err = s.db.QueryRowContext(ctx, `
		SELECT count(*) FILTER (WHERE incident_id = $1),
		       count(*) FILTER (WHERE user_id = $2 AND created_at > now() - make_interval(secs => $3))
		FROM incident_photo WHERE incident_id = $1 OR user_id = $2`, incidentID, userID, window.Seconds()).
		Scan(&byIncident, &byUserRecent)
	return byIncident, byUserRecent, err
}

func (s *Store) GetPhoto(ctx context.Context, id int64) (model.IncidentPhoto, error) {
	var p model.IncidentPhoto
	err := SELECT(IncidentPhoto.ContentType, IncidentPhoto.Data).FROM(IncidentPhoto).
		WHERE(IncidentPhoto.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &p)
	if errors.Is(err, qrm.ErrNoRows) {
		return model.IncidentPhoto{}, ErrNotFound
	}
	return p, err
}

func (s *Store) PhotoIDsByIncident(ctx context.Context, incidentIDs []int64) (map[int64][]int64, error) {
	out := map[int64][]int64{}
	if len(incidentIDs) == 0 {
		return out, nil
	}
	ids := make([]Expression, len(incidentIDs))
	for i, id := range incidentIDs {
		ids[i] = Int64(id)
	}
	var rows []model.IncidentPhoto
	err := SELECT(IncidentPhoto.ID, IncidentPhoto.IncidentID).FROM(IncidentPhoto).
		WHERE(IncidentPhoto.IncidentID.IN(ids...)).ORDER_BY(IncidentPhoto.ID).
		QueryContext(ctx, s.db, &rows)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		return nil, err
	}
	for _, r := range rows {
		if r.IncidentID != nil {
			out[*r.IncidentID] = append(out[*r.IncidentID], r.ID)
		}
	}
	return out, nil
}

func (s *Store) DeleteStalePhotos(ctx context.Context, olderThan time.Duration) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM incident_photo WHERE incident_id IS NULL AND created_at < now() - make_interval(secs => $1)`, olderThan.Seconds())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
