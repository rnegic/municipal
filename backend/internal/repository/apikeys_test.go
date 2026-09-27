package repository

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"ukapp/internal/domain"
)

func seedUkWithDispatcher(t *testing.T, s *Store, ext string) (ukID, userID int64) {
	t.Helper()
	err := s.db.QueryRowContext(context.Background(), `
		WITH org AS (INSERT INTO uk (external_id, name) VALUES ($1, 'УК '||$1) RETURNING id)
		INSERT INTO app_user (full_name, role, uk_id) SELECT 'Диспетчер', 'uk_dispatcher', id FROM org
		RETURNING uk_id, id`, ext).Scan(&ukID, &userID)
	if err != nil {
		t.Fatal(err)
	}
	return ukID, userID
}

func TestApiKeysLifecycle(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	ukID, disp := seedUkWithDispatcher(t, s, "uk-a")
	otherUk, _ := seedUkWithDispatcher(t, s, "uk-b")

	k1, err := s.CreateApiKey(ctx, ukID, disp, "1С офис", "aaaaaaaa", []byte("h1"))
	if err != nil {
		t.Fatal(err)
	}
	k2, err := s.CreateApiKey(ctx, ukID, disp, "1С склад", "bbbbbbbb", []byte("h2"))
	if err != nil {
		t.Fatal(err)
	}
	if k1.UserID != k2.UserID || k1.UserID == disp {
		t.Fatalf("keys of one uk share one service user distinct from creator: %d %d %d", k1.UserID, k2.UserID, disp)
	}
	u, err := s.GetUser(ctx, k1.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != domain.RoleUkDispatcher || u.UkID == nil || *u.UkID != ukID || u.FullName != domain.UkIntegrationUserName || u.PasswordHash != nil || u.AdsAuthority {
		t.Fatalf("service user: %+v", u)
	}

	got, err := s.ApiKeyByHash(ctx, []byte("h1"))
	if err != nil || got.ID != k1.ID {
		t.Fatalf("by hash: %v %+v", err, got)
	}
	if err := s.RevokeApiKey(ctx, otherUk, k1.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoke foreign: %v", err)
	}
	if err := s.RevokeApiKey(ctx, ukID, k1.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.RevokeApiKey(ctx, ukID, k1.ID); err != nil {
		t.Fatalf("repeat revoke must be ok: %v", err)
	}
	if _, err := s.ApiKeyByHash(ctx, []byte("h1")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked key must not authenticate: %v", err)
	}
	keys, err := s.ListApiKeys(ctx, ukID)
	if err != nil || len(keys) != 2 || keys[0].RevokedAt == nil || keys[1].RevokedAt != nil {
		t.Fatalf("list: %v %+v", err, keys)
	}
	if err := s.TouchApiKey(ctx, k2.ID); err != nil {
		t.Fatal(err)
	}
	if k, _ := s.ApiKeyByHash(ctx, []byte("h2")); k.LastUsedAt == nil {
		t.Fatal("touch must set last_used_at")
	}
}

func TestApiKeysLimit(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	ukID, disp := seedUkWithDispatcher(t, s, "uk-a")
	for i := 0; i < domain.UkMaxApiKeys; i++ {
		if _, err := s.CreateApiKey(ctx, ukID, disp, "k", "p", []byte("h"+strconv.Itoa(i))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateApiKey(ctx, ukID, disp, "k", "p", []byte("over")); !errors.Is(err, ErrKeyLimit) {
		t.Fatalf("want ErrKeyLimit, got %v", err)
	}
	keys, _ := s.ListApiKeys(ctx, ukID)
	if err := s.RevokeApiKey(ctx, ukID, keys[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateApiKey(ctx, ukID, disp, "k", "p", []byte("after-revoke")); err != nil {
		t.Fatalf("revoked keys must not count: %v", err)
	}
}
