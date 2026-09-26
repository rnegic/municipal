CREATE TABLE IF NOT EXISTS organization (
  id              TEXT PRIMARY KEY,
  name            TEXT NOT NULL,
  phone           TEXT,
  emergency_phone TEXT,
  email           TEXT,
  website         TEXT,
  office_address  TEXT,
  working_hours   TEXT
);

CREATE TABLE IF NOT EXISTS house (
  id              TEXT PRIMARY KEY,
  fias_id         TEXT NOT NULL UNIQUE,
  address         TEXT NOT NULL,
  organization_id TEXT NOT NULL REFERENCES organization(id)
);

CREATE SEQUENCE IF NOT EXISTS incident_seq;

CREATE TABLE IF NOT EXISTS incident (
  id           TEXT PRIMARY KEY DEFAULT ('INC-' || lpad(nextval('incident_seq')::text, 3, '0')),
  external_ref TEXT NOT NULL UNIQUE,
  house_id     TEXT NOT NULL REFERENCES house(id),
  title        TEXT NOT NULL,
  description  TEXT NOT NULL,
  severity     TEXT NOT NULL CHECK (severity IN ('critical','warning')),
  entrance     TEXT,
  riser        TEXT,
  status       TEXT NOT NULL DEFAULT 'accepted'
               CHECK (status IN ('accepted','in_progress','verifying','done')),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS incident_updated ON incident(updated_at);
ALTER TABLE incident DROP CONSTRAINT IF EXISTS incident_status_check;
ALTER TABLE incident ADD CONSTRAINT incident_status_check
  CHECK (status IN ('accepted','in_progress','verifying','done','false_alarm'));
ALTER TABLE incident ADD COLUMN IF NOT EXISTS suspicious BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS dispatcher (
  login           TEXT PRIMARY KEY,
  password_sha256 TEXT NOT NULL,
  organization_id TEXT NOT NULL REFERENCES organization(id)
);
