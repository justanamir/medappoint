-- Seed: one clinic, provider+user, patient+user, one service, and weekday availability
-- Note: password hash below is bcrypt for plain text "P@ssw0rd!"

-- 1) Clinic
INSERT INTO clinics (name, timezone, address)
VALUES ('Downtown Family Clinic', 'Asia/Kuala_Lumpur', 'Jalan Contoh 123, KL')
ON CONFLICT DO NOTHING;

-- 2) Users (admin, patient, provider) using the same demo password hash
INSERT INTO users (email, password_hash, role)
VALUES
  ('seed-admin@example.com',   '$2a$10$VtH8o2xwFqg2eCxr1qO1Heqf7c0R2qH1kM9sSj4Y0pQ8rXc6y2p8K', 'admin'),
  ('seed-patient@example.com', '$2a$10$VtH8o2xwFqg2eCxr1qO1Heqf7c0R2qH1kM9sSj4Y0pQ8rXc6y2p8K', 'patient'),
  ('seed-doc@example.com',     '$2a$10$VtH8o2xwFqg2eCxr1qO1Heqf7c0R2qH1kM9sSj4Y0pQ8rXc6y2p8K', 'provider')
ON CONFLICT DO NOTHING;

-- 3) Patient profile
INSERT INTO patients (user_id, full_name, phone)
SELECT id, 'Seed Patient', '+60-1234-5678'
FROM users WHERE email='seed-patient@example.com'
ON CONFLICT DO NOTHING;

-- 4) Provider tied to clinic
WITH c AS (SELECT id AS clinic_id FROM clinics WHERE name='Downtown Family Clinic'),
     u AS (SELECT id AS user_id   FROM users   WHERE email='seed-doc@example.com')
INSERT INTO providers (user_id, full_name, speciality, clinic_id)
SELECT u.user_id, 'Dr. Siti Noor', 'General Practice', c.clinic_id
FROM c, u
ON CONFLICT DO NOTHING;

-- 5) One service (30 min)
INSERT INTO services (clinic_id, name, description, duration_min)
SELECT id, 'General Consultation (30m)', 'Standard 30-minute consult', 30
FROM clinics WHERE name='Downtown Family Clinic'
ON CONFLICT DO NOTHING;

-- 6) Availability: Mon–Fri 09:00–17:00 for provider
INSERT INTO availabilities (provider_id, weekday, start_hhmm, end_hhmm)
SELECT p.id, wd, '09:00', '17:00'
FROM providers p
CROSS JOIN (VALUES (1),(2),(3),(4),(5)) AS days(wd)
WHERE p.full_name='Dr. Siti Noor'
ON CONFLICT DO NOTHING;
