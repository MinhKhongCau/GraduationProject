BEGIN;

DROP INDEX IF EXISTS idx_specializations_slug;
DROP INDEX IF EXISTS idx_specializations_code;

ALTER TABLE specializations
    DROP COLUMN location,
    DROP COLUMN symptoms,
    DROP COLUMN slug,
    DROP COLUMN code;

COMMIT;
