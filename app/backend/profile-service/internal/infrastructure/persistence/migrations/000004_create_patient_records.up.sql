-- Hồ sơ người khám: một tài khoản bệnh nhân có thể đặt lịch cho chính mình (SELF)
-- hoặc cho người thân. Hồ sơ SELF được tạo tự động từ thông tin profile khi cần.
BEGIN;

CREATE TABLE IF NOT EXISTS patient_records (
    record_id     uuid         NOT NULL DEFAULT gen_random_uuid(),
    owner_auth_id uuid         NOT NULL,
    full_name     varchar(255) NOT NULL,
    date_of_birth date,
    gender        varchar(20),
    phone_number  varchar(20),
    email         varchar(255),
    address       text,
    relationship  varchar(20)  NOT NULL DEFAULT 'SELF',
    created_at    timestamptz,
    updated_at    timestamptz,
    deleted_at    timestamptz,
    CONSTRAINT patient_records_pkey PRIMARY KEY (record_id)
);

CREATE INDEX IF NOT EXISTS idx_patient_records_owner_auth_id ON patient_records (owner_auth_id);
CREATE INDEX IF NOT EXISTS idx_patient_records_deleted_at ON patient_records (deleted_at);
-- Mỗi tài khoản chỉ có tối đa 1 hồ sơ SELF còn hiệu lực.
CREATE UNIQUE INDEX IF NOT EXISTS idx_patient_records_owner_self
    ON patient_records (owner_auth_id)
    WHERE relationship = 'SELF' AND deleted_at IS NULL;

COMMIT;
