BEGIN;

-- Admin duyệt (VERIFIED) chuyên gia trở thành người quản lý chuyên gia đó.
ALTER TABLE expert_profiles ADD COLUMN IF NOT EXISTS managed_by_admin_id uuid NULL;
ALTER TABLE expert_profiles ADD COLUMN IF NOT EXISTS verified_at timestamptz NULL;
CREATE INDEX IF NOT EXISTS idx_expert_profiles_managed_by_admin_id ON expert_profiles (managed_by_admin_id);

COMMIT;
