-- Enable extensions needed
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Users table
CREATE TABLE IF NOT EXISTS users (
  id            BIGSERIAL PRIMARY KEY,
  email         CITEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role          TEXT NOT NULL CHECK (role IN ('patient','provider','admin')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Patients table
CREATE TABLE IF NOT EXISTS patients (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  full_name   TEXT NOT NULL,
  phone       TEXT,
  dob         DATE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Clinics table
CREATE TABLE IF NOT EXISTS clinics (
  id          BIGSERIAL PRIMARY KEY,
  name        TEXT NOT NULL,
  timezone    TEXT NOT NULL,
  address     TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Providers table
CREATE TABLE IF NOT EXISTS providers (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  full_name   TEXT NOT NULL,
  speciality  TEXT NOT NULL,
  clinic_id   BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Services table
CREATE TABLE IF NOT EXISTS services (
  id           BIGSERIAL PRIMARY KEY,
  clinic_id    BIGINT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  description  TEXT,
  duration_min INTEGER NOT NULL CHECK (duration_min > 0),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Weekly availability (1=Mon ... 7=Sun)
CREATE TABLE IF NOT EXISTS availabilities (
  id           BIGSERIAL PRIMARY KEY,
  provider_id  BIGINT NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
  weekday      INTEGER NOT NULL CHECK (weekday BETWEEN 1 AND 7),
  start_hhmm   TEXT NOT NULL,
  end_hhmm     TEXT NOT NULL
);

-- Blackouts (clinic or provider level)
CREATE TABLE IF NOT EXISTS blackouts (
  id           BIGSERIAL PRIMARY KEY,
  clinic_id    BIGINT REFERENCES clinics(id) ON DELETE CASCADE,
  provider_id  BIGINT REFERENCES providers(id) ON DELETE CASCADE,
  date         DATE NOT NULL,
  reason       TEXT,
  CHECK (clinic_id IS NOT NULL OR provider_id IS NOT NULL)
);

-- Appointments table
CREATE TABLE IF NOT EXISTS appointments (
  id           BIGSERIAL PRIMARY KEY,
  clinic_id    BIGINT NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
  provider_id  BIGINT NOT NULL REFERENCES providers(id) ON DELETE RESTRICT,
  patient_id   BIGINT NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
  service_id   BIGINT NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
  start_time   TIMESTAMPTZ NOT NULL,
  end_time     TIMESTAMPTZ NOT NULL,
  status       TEXT NOT NULL CHECK (status IN ('scheduled','completed','cancelled')),
  notes        TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  -- prevent overlapping for same provider
  EXCLUDE USING gist (
    provider_id WITH =,
    tstzrange(start_time, end_time, '[)') WITH &&
  )
);
