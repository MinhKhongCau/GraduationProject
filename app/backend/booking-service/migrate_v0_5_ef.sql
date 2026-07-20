-- V0.5-EF: reserve slot status 3 for UNAVAILABLE.
-- Existing status values and rows are preserved.
ALTER TABLE "Booking_Expert_Slots"
    DROP CONSTRAINT IF EXISTS chk_booking_expert_slots_status;

ALTER TABLE "Booking_Expert_Slots"
    ADD CONSTRAINT chk_booking_expert_slots_status
    CHECK (status IN (0, 1, 2, 3)) NOT VALID;

ALTER TABLE "Booking_Expert_Slots"
    VALIDATE CONSTRAINT chk_booking_expert_slots_status;
