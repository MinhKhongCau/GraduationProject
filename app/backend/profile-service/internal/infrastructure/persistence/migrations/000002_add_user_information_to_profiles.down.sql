BEGIN;

ALTER TABLE patient_profiles
    ADD COLUMN phone_number  varchar(20),
    ADD COLUMN date_of_birth date,
    ADD COLUMN gender        varchar(20);

ALTER TABLE expert_profiles ADD COLUMN phone_number varchar(20);

UPDATE patient_profiles pp
SET phone_number  = p.phone_number,
    date_of_birth = p.date_of_birth,
    gender        = p.gender
FROM profiles p
WHERE p.id = pp.profile_id;

UPDATE expert_profiles ep
SET phone_number = p.phone_number
FROM profiles p
WHERE p.id = ep.profile_id;

ALTER TABLE profiles
    DROP COLUMN country,
    DROP COLUMN phone_number,
    DROP COLUMN gender,
    DROP COLUMN date_of_birth;

ALTER TABLE profiles RENAME COLUMN full_name TO name;

COMMIT;
