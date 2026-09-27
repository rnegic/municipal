package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/repository"

	"ukapp/gen/db/ukapp/public/model"
)

func seedDispatcherUser(t *testing.T, s *repository.Store) model.AppUser {
	t.Helper()
	var id int64
	err := s.DB().QueryRowContext(context.Background(), `
		WITH org AS (INSERT INTO uk (external_id, name, license_number, license_valid_until)
		             VALUES ('uk-k', 'УК К', '16-000001', '2099-12-31') RETURNING id)
		INSERT INTO app_user (full_name, role, uk_id) SELECT 'Диспетчер', 'uk_dispatcher', id FROM org RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.GetUser(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestApiKeyAuth(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	ctx := context.Background()
	disp := seedDispatcherUser(t, s)

	if _, err := svc.CreateApiKey(ctx, disp, "   "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("blank name: %v", err)
	}
	created, err := svc.CreateApiKey(ctx, disp, "1С")
	if err != nil {
		t.Fatal(err)
	}
	u, err := svc.AuthenticateApiKey(ctx, created.Plain)
	if err != nil || u.ID != created.Key.UserID || *u.UkID != *disp.UkID {
		t.Fatalf("auth: %v %+v", err, u)
	}
	if _, err := svc.AuthenticateApiKey(ctx, created.Plain+"x"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("wrong key: %v", err)
	}
	if err := svc.RevokeApiKey(ctx, disp, created.Key.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AuthenticateApiKey(ctx, created.Plain); !errors.Is(err, ErrForbidden) {
		t.Fatalf("revoked key: %v", err)
	}
}

func TestKeyLimiter(t *testing.T) {
	l := newKeyLimiter(domain.UkApiKeyRPS)
	now := time.Unix(100, 0)
	for i := 0; i < domain.UkApiKeyRPS; i++ {
		if !l.allow(1, now) {
			t.Fatalf("request %d must pass", i)
		}
	}
	if l.allow(1, now.Add(500*time.Millisecond)) {
		t.Fatal("over limit within the second must fail")
	}
	if !l.allow(2, now) {
		t.Fatal("other key has own budget")
	}
	if !l.allow(1, now.Add(time.Second)) {
		t.Fatal("next window must pass")
	}
}

func TestApiKeyAuth_LicenseExpired(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	ctx := context.Background()
	disp := seedDispatcherUser(t, s)
	created, err := svc.CreateApiKey(ctx, disp, "1С")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().ExecContext(ctx, `UPDATE uk SET license_valid_until = '2020-01-01' WHERE id = $1`, *disp.UkID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AuthenticateApiKey(ctx, created.Plain); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expired license must reject key: %v", err)
	}
}
