BEGIN;

DROP INDEX IF EXISTS idx_expert_profiles_managed_by_admin_id;
ALTER TABLE expert_profiles DROP COLUMN IF EXISTS verified_at;
ALTER TABLE expert_profiles DROP COLUMN IF EXISTS managed_by_admin_id;

COMMIT;
