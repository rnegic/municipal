// Package store — Postgres mock-системы УК: реестр организаций, домов, инцидентов, диспетчеров.
// Строковый SQL через database/sql: это не продукт, а референсная реализация контракта.
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // driver "pgx"
)

//go:embed schema.sql
var schemaSQL string

//go:embed seed.sql
var seedSQL string

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store struct{ db *sql.DB }

type House struct {
	ID, FiasID, Address, OrgID, OrgName               string
	OrgPhone, OrgEmergencyPhone, OrgEmail, OrgWebsite *string
	OrgOfficeAddress, OrgWorkingHours                 *string
}

type Incident struct {
	ID, ExternalRef, HouseID, Title, Description, Severity, Status string
	Entrance, Riser                                                *string
	Suspicious                                                     bool
	CreatedAt, UpdatedAt                                           time.Time
	Address                                                        string // house.address, для ЛК
}

type Dispatcher struct{ Login, PasswordSHA256, OrgID string }

// Open connects, applies schema.sql and seed.sql (both idempotent).
func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	for _, q := range []string{schemaSQL, seedSQL} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply sql: %w", err)
		}
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() { _ = s.db.Close() }

// DB exposes the handle for test setup.
func (s *Store) DB() *sql.DB { return s.db }

func (s *Store) FindHouseByFias(ctx context.Context, fias string) (House, error) {
	var h House
	err := s.db.QueryRowContext(ctx,
		`SELECT h.id, h.fias_id, h.address, o.id, o.name,
		        o.phone, o.emergency_phone, o.email, o.website, o.office_address, o.working_hours
		 FROM house h JOIN organization o ON o.id = h.organization_id WHERE h.fias_id = $1`, fias).
		Scan(&h.ID, &h.FiasID, &h.Address, &h.OrgID, &h.OrgName,
			&h.OrgPhone, &h.OrgEmergencyPhone, &h.OrgEmail, &h.OrgWebsite, &h.OrgOfficeAddress, &h.OrgWorkingHours)
	if errors.Is(err, sql.ErrNoRows) {
		return House{}, ErrNotFound
	}
	return h, err
}

const incidentCols = `i.id, i.external_ref, i.house_id, i.title, i.description, i.severity, i.entrance, i.riser, i.status, i.suspicious, i.created_at, i.updated_at, h.address`

func scanIncident(row interface{ Scan(...any) error }) (Incident, error) {
	var in Incident
	err := row.Scan(&in.ID, &in.ExternalRef, &in.HouseID, &in.Title, &in.Description, &in.Severity, &in.Entrance, &in.Riser, &in.Status, &in.Suspicious, &in.CreatedAt, &in.UpdatedAt, &in.Address)
	return in, err
}

func (s *Store) get(ctx context.Context, where string, arg any) (Incident, error) {
	in, err := scanIncident(s.db.QueryRowContext(ctx, `SELECT `+incidentCols+` FROM incident i JOIN house h ON h.id = i.house_id WHERE `+where, arg))
	if errors.Is(err, sql.ErrNoRows) {
		return Incident{}, ErrNotFound
	}
	return in, err
}

// CreateIncident is idempotent by external_ref: created=false returns the existing row.
// Unknown house_id → ErrNotFound.
func (s *Store) CreateIncident(ctx context.Context, in Incident) (Incident, bool, error) {
	if existing, err := s.get(ctx, `i.external_ref = $1`, in.ExternalRef); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Incident{}, false, err
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM house WHERE id = $1)`, in.HouseID).Scan(&exists); err != nil {
		return Incident{}, false, err
	}
	if !exists {
		return Incident{}, false, ErrNotFound
	}
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO incident (external_ref, house_id, title, description, severity, entrance, riser, suspicious)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		in.ExternalRef, in.HouseID, in.Title, in.Description, in.Severity, in.Entrance, in.Riser, in.Suspicious).Scan(&id)
	if err != nil {
		return Incident{}, false, err
	}
	created, err := s.get(ctx, `i.id = $1`, id)
	return created, true, err
}

func (s *Store) UpdatedSince(ctx context.Context, since time.Time) ([]Incident, error) {
	return s.list(ctx, `i.updated_at > $1 ORDER BY i.updated_at`, since)
}

func (s *Store) ListByOrg(ctx context.Context, orgID string) ([]Incident, error) {
	return s.list(ctx, `h.organization_id = $1 ORDER BY i.created_at DESC`, orgID)
}

func (s *Store) list(ctx context.Context, whereOrder string, arg any) ([]Incident, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+incidentCols+` FROM incident i JOIN house h ON h.id = i.house_id WHERE `+whereOrder, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Incident{}
	for rows.Next() {
		in, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, in)
	}
	return out, rows.Err()
}

// SetStatus: ErrNotFound for unknown id, ErrConflict when leaving done or false_alarm.
func (s *Store) SetStatus(ctx context.Context, id, status string) (Incident, error) {
	cur, err := s.get(ctx, `i.id = $1`, id)
	if err != nil {
		return Incident{}, err
	}
	if (cur.Status == "done" || cur.Status == "false_alarm") && status != cur.Status {
		return Incident{}, ErrConflict
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE incident SET status = $2, updated_at = now() WHERE id = $1 AND status <> $2`, id, status); err != nil {
		return Incident{}, err
	}
	return s.get(ctx, `i.id = $1`, id)
}

func (s *Store) Dispatcher(ctx context.Context, login string) (Dispatcher, error) {
	var d Dispatcher
	err := s.db.QueryRowContext(ctx, `SELECT login, password_sha256, organization_id FROM dispatcher WHERE login = $1`, login).
		Scan(&d.Login, &d.PasswordSHA256, &d.OrgID)
	if errors.Is(err, sql.ErrNoRows) {
		return Dispatcher{}, ErrNotFound
	}
	return d, err
}
