-- Remove availability for the seeded provider
DELETE FROM availabilities
WHERE provider_id IN (SELECT id FROM providers WHERE full_name='Dr. Siti Noor');

-- Remove service
DELETE FROM services
WHERE name='General Consultation (30m)'
  AND clinic_id IN (SELECT id FROM clinics WHERE name='Downtown Family Clinic');

-- Remove provider
DELETE FROM providers
WHERE full_name='Dr. Siti Noor'
  AND clinic_id IN (SELECT id FROM clinics WHERE name='Downtown Family Clinic');

-- Remove patient profile
DELETE FROM patients
WHERE user_id IN (SELECT id FROM users WHERE email='seed-patient@example.com');

-- Remove the three seed users
DELETE FROM users
WHERE email IN ('seed-admin@example.com','seed-patient@example.com','seed-doc@example.com');

-- Remove the clinic
DELETE FROM clinics WHERE name='Downtown Family Clinic';
