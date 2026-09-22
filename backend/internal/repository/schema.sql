CREATE TABLE IF NOT EXISTS uk (
  id          BIGSERIAL PRIMARY KEY,
  external_id TEXT NOT NULL UNIQUE,   -- id организации в системе УК (contracts/uk.yaml)
  name        TEXT NOT NULL,
  rating      INT  NOT NULL DEFAULT 0
);

ALTER TABLE uk ADD COLUMN IF NOT EXISTS external_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_id_key ON uk(external_id);

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
  max_user_id           BIGINT NOT NULL UNIQUE,
  full_name             TEXT NOT NULL,
  role                  TEXT NOT NULL DEFAULT 'resident' CHECK (role IN ('resident')),
  house_id              BIGINT REFERENCES house(id),
  uk_id                 BIGINT REFERENCES uk(id),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

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

CREATE TABLE IF NOT EXISTS incident_photo (
  id           BIGSERIAL PRIMARY KEY,
  incident_id  BIGINT NOT NULL REFERENCES incident(id),
  user_id      BIGINT NOT NULL REFERENCES app_user(id),
  content_type TEXT NOT NULL,
  data         BYTEA NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS incident_photo_incident ON incident_photo(incident_id);

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
