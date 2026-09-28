package repository

import (
	"context"
	"database/sql"
	"errors"

	. "github.com/go-jet/jet/v2/postgres"

	"ukapp/internal/domain"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

var ErrKeyLimit = errors.New("api key limit reached")

func (s *Store) CreateApiKey(ctx context.Context, ukID, createdBy int64, name, prefix string, hash []byte) (model.UkAPIKey, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.UkAPIKey{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT id FROM uk WHERE id = $1 FOR UPDATE`, ukID); err != nil {
		return model.UkAPIKey{}, err
	}
	var active int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM uk_api_key WHERE uk_id = $1 AND revoked_at IS NULL`, ukID).Scan(&active); err != nil {
		return model.UkAPIKey{}, err
	}
	if active >= domain.UkMaxApiKeys {
		return model.UkAPIKey{}, ErrKeyLimit
	}
	var userID int64
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM uk_api_key WHERE uk_id = $1 LIMIT 1`, ukID).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO app_user (full_name, role, uk_id) VALUES ($1, $2, $3) RETURNING id`,
			domain.UkIntegrationUserName, domain.RoleUkDispatcher, ukID).Scan(&userID)
	}
	if err != nil {
		return model.UkAPIKey{}, err
	}
	var out model.UkAPIKey
	err = UkAPIKey.INSERT(UkAPIKey.UkID, UkAPIKey.UserID, UkAPIKey.Name, UkAPIKey.Prefix, UkAPIKey.KeyHash, UkAPIKey.CreatedBy).
		VALUES(ukID, userID, name, prefix, hash, createdBy).
		RETURNING(UkAPIKey.AllColumns).QueryContext(ctx, tx, &out)
	if err != nil {
		return model.UkAPIKey{}, err
	}
	return out, tx.Commit()
}

func (s *Store) ListApiKeys(ctx context.Context, ukID int64) ([]model.UkAPIKey, error) {
	out := []model.UkAPIKey{}
	err := SELECT(UkAPIKey.AllColumns).FROM(UkAPIKey).
		WHERE(UkAPIKey.UkID.EQ(Int64(ukID))).ORDER_BY(UkAPIKey.ID.ASC()).
		QueryContext(ctx, s.db, &out)
	return out, err
}

func (s *Store) RevokeApiKey(ctx context.Context, ukID, id int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE uk_api_key SET revoked_at = COALESCE(revoked_at, now()) WHERE id = $1 AND uk_id = $2`, id, ukID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ApiKeyByHash(ctx context.Context, hash []byte) (model.UkAPIKey, error) {
	var out model.UkAPIKey
	err := SELECT(UkAPIKey.AllColumns).FROM(UkAPIKey).
		WHERE(UkAPIKey.KeyHash.EQ(Bytea(hash)).AND(UkAPIKey.RevokedAt.IS_NULL())).
		QueryContext(ctx, s.db, &out)
	return out, notFound(err)
}

func (s *Store) TouchApiKey(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE uk_api_key SET last_used_at = now() WHERE id = $1`, id)
	return err
}
