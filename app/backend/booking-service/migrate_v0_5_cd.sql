-- V0.5-CD: source provenance for safe future-slot reconciliation.
-- Existing rows intentionally remain NULL; no destructive backfill is attempted.
ALTER TABLE "Booking_Expert_Slots"
    ADD COLUMN IF NOT EXISTS availability_id uuid;

CREATE INDEX IF NOT EXISTS idx_booking_expert_slots_availability_id
    ON "Booking_Expert_Slots" (availability_id);
