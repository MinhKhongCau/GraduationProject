-- V1: one active payment attempt per appointment and immutable payment expiry.
BEGIN;

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS expires_at BIGINT;

ALTER TABLE payment_orders
    DROP CONSTRAINT IF EXISTS payment_orders_status_check;

ALTER TABLE payment_orders
    DROP CONSTRAINT IF EXISTS chk_payment_orders_status;

ALTER TABLE payment_orders
    ADD CONSTRAINT chk_payment_orders_status
    CHECK (status IN (1, 2, 3, 4)) NOT VALID;

ALTER TABLE payment_orders
    VALIDATE CONSTRAINT chk_payment_orders_status;

-- Legacy pending attempts have no trustworthy booking-lock expiry. Preserve them
-- for audit but make them non-payable before enforcing active-order uniqueness.
UPDATE payment_orders
SET status = 4
WHERE status = 1
  AND expires_at IS NULL;

-- A partially deployed V1 application may have written multiple non-null
-- PENDING attempts before the unique index existed. Expire every ambiguous
-- attempt so the next request must revalidate booking eligibility.
UPDATE payment_orders
SET status = 4
WHERE status = 1
  AND appointment_id IN (
      SELECT appointment_id
      FROM payment_orders
      WHERE appointment_id IS NOT NULL
        AND status = 1
      GROUP BY appointment_id
      HAVING COUNT(*) > 1
  );

-- Successful payment history is authoritative and must never be rewritten.
-- Stop with an actionable error instead of silently deleting or converting it.
DO $$
DECLARE
    duplicate_count BIGINT;
    duplicate_appointments TEXT;
BEGIN
    SELECT COUNT(*), STRING_AGG(appointment_id::text, ', ' ORDER BY appointment_id::text)
    INTO duplicate_count, duplicate_appointments
    FROM (
        SELECT appointment_id
        FROM payment_orders
        WHERE appointment_id IS NOT NULL
          AND status = 2
        GROUP BY appointment_id
        HAVING COUNT(*) > 1
    ) duplicate_success;

    IF duplicate_count > 0 THEN
        RAISE EXCEPTION
            'cannot create unique SUCCESS payment index: % appointment(s) have duplicate SUCCESS orders: %',
            duplicate_count,
            duplicate_appointments;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS ix_payment_orders_appointment_id
    ON payment_orders (appointment_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_payment_orders_active_pending_appointment
    ON payment_orders (appointment_id)
    WHERE appointment_id IS NOT NULL AND status = 1;

CREATE UNIQUE INDEX IF NOT EXISTS ux_payment_orders_success_appointment
    ON payment_orders (appointment_id)
    WHERE appointment_id IS NOT NULL AND status = 2;

ALTER TABLE payment_orders
    DROP CONSTRAINT IF EXISTS chk_payment_orders_pending_expiry;

ALTER TABLE payment_orders
    ADD CONSTRAINT chk_payment_orders_pending_expiry
    CHECK (status <> 1 OR (expires_at IS NOT NULL AND expires_at > 0)) NOT VALID;

ALTER TABLE payment_orders
    VALIDATE CONSTRAINT chk_payment_orders_pending_expiry;

COMMIT;
