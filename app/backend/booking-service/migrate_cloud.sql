-- =======================================================================
-- MIGRATION: Chuyển cột status từ varchar sang smallint
-- Booking Service - Chạy 1 lần trên Cloud DB trước khi deploy code mới
-- =======================================================================
-- Lý do: Code Go dùng enum int (0=AVAILABLE, 1=LOCKED, 2=OCCUPIED),
-- nhưng bảng cũ trên Cloud đang lưu varchar ("AVAILABLE", "LOCKED"...).
-- AutoMigrate sẽ không tự cast được, cần chạy thủ công.
-- =======================================================================

BEGIN;

-- 1. Chuyển cột status trong bảng Slots
ALTER TABLE "Booking_Expert_Slots"
  ALTER COLUMN status TYPE smallint
  USING CASE status::text
    WHEN 'AVAILABLE' THEN 0
    WHEN 'LOCKED'    THEN 1
    WHEN 'OCCUPIED'  THEN 2
	WHEN 'UNAVAILABLE' THEN 3
	WHEN '0' THEN 0
	WHEN '1' THEN 1
	WHEN '2' THEN 2
	WHEN '3' THEN 3
    ELSE 0
  END;

ALTER TABLE "Booking_Expert_Slots"
  ALTER COLUMN status SET DEFAULT 0;

-- 2. Chuyển cột status trong bảng Appointments
ALTER TABLE "Booking_Appointments"
  ALTER COLUMN status TYPE smallint
  USING CASE status::text
    WHEN 'PENDING_PAYMENT' THEN 0
    WHEN 'CONFIRMED'       THEN 1
    WHEN 'CANCELLED'       THEN 2
    ELSE 0
  END;

ALTER TABLE "Booking_Appointments"
  ALTER COLUMN status SET DEFAULT 0;

-- 3. Chuyển cột start_datetime và end_datetime trong bảng Time_Off
--    (Lỗi thứ 2 trong log: timestamp with time zone -> bigint)
ALTER TABLE "Booking_Expert_Time_Off"
  ALTER COLUMN start_datetime TYPE bigint
  USING EXTRACT(EPOCH FROM start_datetime)::bigint * 1000;

ALTER TABLE "Booking_Expert_Time_Off"
  ALTER COLUMN end_datetime TYPE bigint
  USING EXTRACT(EPOCH FROM end_datetime)::bigint * 1000;

COMMIT;

-- Kiểm tra lại sau khi chạy:
-- SELECT slot_id, status FROM "Booking_Expert_Slots" LIMIT 5;
-- SELECT appointment_id, status FROM "Booking_Appointments" LIMIT 5;
