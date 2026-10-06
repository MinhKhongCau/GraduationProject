-- Baseline: schema profile-service đang có (trước đây do GORM AutoMigrate tạo).
-- Dùng IF NOT EXISTS để chạy an toàn trên DB đã được AutoMigrate tạo sẵn: khi đó
-- migration này là no-op và chỉ đánh dấu version 1 trong schema_migrations.

CREATE TABLE IF NOT EXISTS profiles (
    id         uuid         NOT NULL DEFAULT gen_random_uuid(),
    slug       varchar(255) NOT NULL,
    name       varchar(255) NOT NULL,
    auth_id    uuid         NOT NULL,
    role       varchar(20)  NOT NULL,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz,
    CONSTRAINT profiles_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_slug ON profiles (slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_auth_id ON profiles (auth_id);
CREATE INDEX IF NOT EXISTS idx_profiles_role ON profiles (role);
CREATE INDEX IF NOT EXISTS idx_profiles_deleted_at ON profiles (deleted_at);

CREATE TABLE IF NOT EXISTS patient_profiles (
    profile_id    uuid NOT NULL,
    phone_number  varchar(20),
    email         varchar(255),
    avatar_url    text,
    date_of_birth date,
    gender        varchar(20),
    address       text,
    CONSTRAINT patient_profiles_pkey PRIMARY KEY (profile_id),
    CONSTRAINT fk_profiles_patient_profile FOREIGN KEY (profile_id) REFERENCES profiles (id)
);

CREATE TABLE IF NOT EXISTS medical_histories (
    history_id         uuid         NOT NULL DEFAULT gen_random_uuid(),
    patient_profile_id uuid         NOT NULL,
    condition_name     varchar(255) NOT NULL,
    description        text,
    diagnosed_at       date,
    is_chronic         boolean DEFAULT false,
    is_active          boolean DEFAULT true,
    CONSTRAINT medical_histories_pkey PRIMARY KEY (history_id),
    CONSTRAINT fk_patient_profiles_medical_histories FOREIGN KEY (patient_profile_id) REFERENCES patient_profiles (profile_id)
);
CREATE INDEX IF NOT EXISTS idx_medical_histories_patient_profile_id ON medical_histories (patient_profile_id);

CREATE TABLE IF NOT EXISTS expert_profiles (
    profile_id             uuid NOT NULL,
    phone_number           varchar(20),
    email                  varchar(255),
    avatar_url             text,
    introduction_video_url text,
    bio                    text,
    verification_status    varchar(50) DEFAULT 'UNVERIFIED',
    CONSTRAINT expert_profiles_pkey PRIMARY KEY (profile_id),
    CONSTRAINT fk_profiles_expert_profile FOREIGN KEY (profile_id) REFERENCES profiles (id)
);

CREATE TABLE IF NOT EXISTS specializations (
    spec_id     uuid         NOT NULL DEFAULT gen_random_uuid(),
    name        varchar(255) NOT NULL,
    description text,
    image_url   text,
    is_active   boolean DEFAULT true,
    CONSTRAINT specializations_pkey PRIMARY KEY (spec_id)
);

CREATE TABLE IF NOT EXISTS expert_specializations (
    expert_profile_id uuid NOT NULL,
    spec_id           uuid NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT expert_specializations_pkey PRIMARY KEY (expert_profile_id, spec_id),
    CONSTRAINT fk_expert_specializations_expert_profile FOREIGN KEY (expert_profile_id) REFERENCES expert_profiles (profile_id),
    CONSTRAINT fk_expert_specializations_specialization FOREIGN KEY (spec_id) REFERENCES specializations (spec_id)
);

CREATE TABLE IF NOT EXISTS admin_profiles (
    profile_id uuid NOT NULL,
    email      varchar(255),
    note       text,
    CONSTRAINT admin_profiles_pkey PRIMARY KEY (profile_id),
    CONSTRAINT fk_profiles_admin_profile FOREIGN KEY (profile_id) REFERENCES profiles (id)
);
