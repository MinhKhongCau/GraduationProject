BEGIN;

ALTER TABLE "Booking_Config_Availability"
  ADD COLUMN IF NOT EXISTS price numeric(12,2);

COMMIT;

-- Existing availability rows intentionally remain NULL. Experts must set an
-- explicit whole-VND consultation price before those rows can generate slots.
