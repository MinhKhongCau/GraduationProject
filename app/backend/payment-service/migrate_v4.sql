-- V4: transaction management per role. Admin chỉ quản lý giao dịch của chuyên gia mình đã duyệt
-- (phạm vi lấy từ profile-service), nên truy vấn theo expert_id/payer_id cần index; Admin có thể
-- đưa đơn vào diện kiểm tra/hoàn tiền (ADMIN_REVIEW) và đóng hồ sơ bồi hoàn (RESOLVED).
BEGIN;

CREATE INDEX IF NOT EXISTS ix_payment_orders_expert_created
    ON payment_orders (expert_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_payment_orders_payer_created
    ON payment_orders (payer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_payment_wallets_user_id
    ON payment_wallets (user_id);
CREATE INDEX IF NOT EXISTS ix_payment_wallet_transactions_wallet_created
    ON payment_wallet_transactions (wallet_id, created_at DESC);

ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_type;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_type
    CHECK (type IN ('BOOKING_FULFILLMENT', 'DUPLICATE_GATEWAY_CAPTURE', 'ADMIN_REVIEW'));

ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_status;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_status
    CHECK (status IN ('MANUAL_REVIEW', 'REFUND_REQUIRED', 'RESOLVED'));

ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_reason;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_reason
    CHECK (reason_code IN (
        'BOOKING_CONFLICT',
        'APPOINTMENT_NOT_FOUND',
        'BOOKING_AUTHENTICATION_REJECTED',
        'BOOKING_DELIVERY_RETRIES_EXHAUSTED',
        'BOOKING_CONTRACT_FAILURE',
        'DUPLICATE_GATEWAY_CAPTURE',
        'ADMIN_MANUAL_REVIEW',
        'ADMIN_REFUND_REQUEST'
    ));

-- Hồ sơ đã đóng phải có thời điểm đóng.
ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_resolved_time;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_resolved_time
    CHECK (status <> 'RESOLVED' OR resolved_at IS NOT NULL);

COMMIT;
