CREATE TABLE IF NOT EXISTS uk (
  id          BIGSERIAL PRIMARY KEY,
  external_id TEXT NOT NULL UNIQUE,   -- id организации в системе УК (contracts/uk.yaml)
  name        TEXT NOT NULL,
  rating      INT  NOT NULL DEFAULT 0
);

ALTER TABLE uk ADD COLUMN IF NOT EXISTS external_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_id_key ON uk(external_id);
ALTER TABLE uk ADD COLUMN IF NOT EXISTS inn TEXT;
ALTER TABLE uk ADD COLUMN IF NOT EXISTS ogrn TEXT;
ALTER TABLE uk ADD COLUMN IF NOT EXISTS license_number TEXT;
ALTER TABLE uk ADD COLUMN IF NOT EXISTS license_valid_until DATE;
CREATE UNIQUE INDEX IF NOT EXISTS uk_inn_key ON uk(inn);

CREATE TABLE IF NOT EXISTS house (
  id              BIGSERIAL PRIMARY KEY,
  address_raw     TEXT NOT NULL,
  house_fias_id   TEXT NOT NULL UNIQUE,
  uk_id           BIGINT NOT NULL REFERENCES uk(id),
  external_id     TEXT,                -- id дома в системе УК
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS app_user (
  id                    BIGSERIAL PRIMARY KEY,
  max_user_id           BIGINT UNIQUE,
  full_name             TEXT NOT NULL,
  role                  TEXT NOT NULL DEFAULT 'resident',
  house_id              BIGINT REFERENCES house(id),
  uk_id                 BIGINT REFERENCES uk(id),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE app_user ALTER COLUMN max_user_id DROP NOT NULL;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS position TEXT;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS password_hash TEXT;
ALTER TABLE app_user ADD COLUMN IF NOT EXISTS ads_authority BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE app_user DROP CONSTRAINT IF EXISTS app_user_role_check;
ALTER TABLE app_user ADD CONSTRAINT app_user_role_check CHECK (
  (role = 'resident' AND max_user_id IS NOT NULL) OR (role = 'uk_dispatcher' AND uk_id IS NOT NULL));

CREATE TABLE IF NOT EXISTS uk_login_attempt (
  id          BIGSERIAL PRIMARY KEY,
  inn         TEXT NOT NULL,
  max_user_id BIGINT,
  user_id     BIGINT REFERENCES app_user(id),
  success     BOOLEAN NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS uk_login_attempt_failures ON uk_login_attempt(inn, created_at) WHERE NOT success;

CREATE TABLE IF NOT EXISTS incident (
  id           BIGSERIAL PRIMARY KEY,
  house_id     BIGINT NOT NULL REFERENCES house(id),
  title        TEXT NOT NULL,
  severity     TEXT NOT NULL CHECK (severity IN ('critical','warning')),
  reporter_id  BIGINT NOT NULL REFERENCES app_user(id),
  description  TEXT NOT NULL,
  entrance     TEXT,
  riser        TEXT,
  status       TEXT NOT NULL DEFAULT 'accepted'
               CHECK (status IN ('accepted','in_progress','verifying','done')),
  external_id  TEXT UNIQUE,            -- id в системе УК; NULL = ещё не зарегистрирован
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  due_at       TIMESTAMPTZ,
  resolved_at  TIMESTAMPTZ
);
ALTER TABLE incident ADD COLUMN IF NOT EXISTS due_at TIMESTAMPTZ;
ALTER TABLE incident ADD COLUMN IF NOT EXISTS external_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS incident_external_id_key ON incident(external_id);
CREATE INDEX IF NOT EXISTS incident_house_open ON incident(house_id, title, created_at)
  WHERE status IN ('accepted','in_progress');
CREATE INDEX IF NOT EXISTS incident_house_status ON incident(house_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS incident_reporter ON incident(house_id, reporter_id, created_at DESC);
CREATE INDEX IF NOT EXISTS incident_unregistered ON incident(id) WHERE external_id IS NULL;

CREATE TABLE IF NOT EXISTS incident_subscription (
  incident_id BIGINT NOT NULL REFERENCES incident(id),
  user_id     BIGINT NOT NULL REFERENCES app_user(id),
  joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (incident_id, user_id)
);

-- Одно подтверждение «починили» от жителя = 1 строка (идемпотентно, см. confirm-эндпоинт).
CREATE TABLE IF NOT EXISTS incident_confirmation (
  incident_id  BIGINT NOT NULL REFERENCES incident(id),
  user_id      BIGINT NOT NULL REFERENCES app_user(id),
  confirmed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (incident_id, user_id)
);

CREATE TABLE IF NOT EXISTS incident_report (
  id                BIGSERIAL PRIMARY KEY,
  incident_id       BIGINT NOT NULL REFERENCES incident(id),
  reporter_id       BIGINT NOT NULL REFERENCES app_user(id),
  house_id          BIGINT NOT NULL,
  title             TEXT NOT NULL,
  description       TEXT NOT NULL,
  severity          TEXT NOT NULL CHECK (severity IN ('critical','warning')),
  entrance          TEXT,
  riser             TEXT,
  outcome           TEXT NOT NULL CHECK (outcome IN ('created','joined')),
  dedup_version     TEXT NOT NULL,
  source_request_id UUID,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (incident_id, reporter_id)
);

CREATE TABLE IF NOT EXISTS uk_event (
  id             BIGSERIAL PRIMARY KEY,
  house_id       BIGINT NOT NULL REFERENCES house(id),
  author_id      BIGINT NOT NULL REFERENCES app_user(id),
  reason         TEXT NOT NULL,
  responsible    TEXT NOT NULL,
  entrance       TEXT,
  riser          TEXT,
  scheduled_from TIMESTAMPTZ NOT NULL,
  scheduled_to   TIMESTAMPTZ NOT NULL CHECK (scheduled_to > scheduled_from),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS uk_event_house ON uk_event(house_id, scheduled_to);

CREATE TABLE IF NOT EXISTS incident_photo (
  id           BIGSERIAL PRIMARY KEY,
  incident_id  BIGINT NOT NULL REFERENCES incident(id),
  user_id      BIGINT NOT NULL REFERENCES app_user(id),
  content_type TEXT NOT NULL,
  data         BYTEA NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS incident_photo_incident ON incident_photo(incident_id);
ALTER TABLE incident_photo ALTER COLUMN incident_id DROP NOT NULL;
CREATE INDEX IF NOT EXISTS incident_photo_staged ON incident_photo(created_at) WHERE incident_id IS NULL;

CREATE TABLE IF NOT EXISTS outbox_message (
  id                 BIGSERIAL PRIMARY KEY,
  target_max_user_id BIGINT NOT NULL,
  kind               TEXT NOT NULL,
  payload            JSONB NOT NULL,
  status             TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','failed')),
  attempts           INT NOT NULL DEFAULT 0,
  next_attempt_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  sent_at            TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS outbox_pending ON outbox_message(next_attempt_at) WHERE status = 'pending';
