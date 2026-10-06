-- User information (full name, date of birth, gender, phone number, country) dùng chung
-- mọi vai trò, chuyển lên bảng profiles. Dữ liệu cũ ở patient_profiles / expert_profiles
-- được copy sang trước khi xoá cột để không còn hai nguồn dữ liệu.
BEGIN;

ALTER TABLE profiles RENAME COLUMN name TO full_name;

ALTER TABLE profiles
    ADD COLUMN date_of_birth date,
    ADD COLUMN gender        varchar(20),
    ADD COLUMN phone_number  varchar(20),
    ADD COLUMN country       varchar(100);

UPDATE profiles p
SET date_of_birth = pp.date_of_birth,
    gender        = NULLIF(pp.gender, ''),
    phone_number  = NULLIF(pp.phone_number, '')
FROM patient_profiles pp
WHERE pp.profile_id = p.id;

UPDATE profiles p
SET phone_number = NULLIF(ep.phone_number, '')
FROM expert_profiles ep
WHERE ep.profile_id = p.id;

ALTER TABLE patient_profiles
    DROP COLUMN phone_number,
    DROP COLUMN date_of_birth,
    DROP COLUMN gender;

ALTER TABLE expert_profiles DROP COLUMN phone_number;

COMMIT;
