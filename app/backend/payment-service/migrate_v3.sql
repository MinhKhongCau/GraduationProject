-- V3: separate payment gateway truth from booking fulfillment and compensation.
BEGIN;

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS fulfillment_status VARCHAR(30),
    ADD COLUMN IF NOT EXISTS gateway_capture_status VARCHAR(30),
    ADD COLUMN IF NOT EXISTS gateway_response_code VARCHAR(10),
    ADD COLUMN IF NOT EXISTS gateway_transaction_status VARCHAR(10),
    ADD COLUMN IF NOT EXISTS gateway_payment_date VARCHAR(14);

ALTER TABLE payment_outbox_events
    ADD COLUMN IF NOT EXISTS terminal_reason_code VARCHAR(50);

CREATE TABLE IF NOT EXISTS payment_compensation_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_order_id UUID NOT NULL,
    appointment_id UUID NOT NULL,
    type VARCHAR(40) NOT NULL,
    status VARCHAR(30) NOT NULL,
    reason_code VARCHAR(60) NOT NULL,
    safe_reason VARCHAR(500) NOT NULL,
    gateway_order_reference VARCHAR(255),
    gateway_transaction_number VARCHAR(255),
    gateway_response_code VARCHAR(10),
    gateway_transaction_status VARCHAR(10),
    gateway_payment_date VARCHAR(14),
    amount_vnd BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    resolved_at BIGINT,
    resolution_note VARCHAR(500)
);

ALTER TABLE payment_compensation_cases
    ADD COLUMN IF NOT EXISTS gateway_order_reference VARCHAR(255),
    ADD COLUMN IF NOT EXISTS gateway_transaction_number VARCHAR(255),
    ADD COLUMN IF NOT EXISTS gateway_response_code VARCHAR(10),
    ADD COLUMN IF NOT EXISTS gateway_transaction_status VARCHAR(10),
    ADD COLUMN IF NOT EXISTS gateway_payment_date VARCHAR(14);

UPDATE payment_compensation_cases AS compensation
SET gateway_order_reference = compensation.payment_order_id::text
WHERE gateway_order_reference IS NULL OR gateway_order_reference = '';

UPDATE payment_compensation_cases AS compensation
SET gateway_transaction_number = COALESCE(compensation.gateway_transaction_number, payment.gateway_txn_ref),
    gateway_response_code = COALESCE(compensation.gateway_response_code, payment.gateway_response_code),
    gateway_transaction_status = COALESCE(compensation.gateway_transaction_status, payment.gateway_transaction_status),
    gateway_payment_date = COALESCE(compensation.gateway_payment_date, payment.gateway_payment_date)
FROM payment_orders AS payment
WHERE payment.id = compensation.payment_order_id;

DO $$
DECLARE
    duplicate_count BIGINT;
BEGIN
    SELECT COUNT(*)
    INTO duplicate_count
    FROM (
        SELECT payment_order_id, reason_code
        FROM payment_compensation_cases
        GROUP BY payment_order_id, reason_code
        HAVING COUNT(*) > 1
    ) duplicate_cases;

    IF duplicate_count > 0 THEN
        RAISE EXCEPTION
            'cannot create compensation idempotency index: % duplicate payment-order/reason group(s) require review',
            duplicate_count;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_payment_compensation_cases_order_reason
    ON payment_compensation_cases (payment_order_id, reason_code);

UPDATE payment_orders
SET gateway_capture_status = CASE
    WHEN status = 2 THEN 'CAPTURED'
    WHEN status = 3 THEN 'FAILED'
    ELSE 'PENDING'
END
WHERE gateway_capture_status IS NULL;

UPDATE payment_orders
SET fulfillment_status = 'PENDING'
WHERE fulfillment_status IS NULL;

-- A delivered booking event is the only safe positive fulfillment evidence.
UPDATE payment_orders AS payment
SET fulfillment_status = 'BOOKING_CONFIRMED'
WHERE payment.status = 2
  AND EXISTS (
      SELECT 1
      FROM payment_outbox_events AS event
      WHERE event.aggregate_type = 'PAYMENT_ORDER'
        AND event.aggregate_id = payment.id
        AND event.event_type = 'booking.appointment.confirm'
        AND event.status = 'DELIVERED'
  );

UPDATE payment_orders AS payment
SET fulfillment_status = 'BOOKING_FAILED'
WHERE payment.status = 3
  AND EXISTS (
      SELECT 1
      FROM payment_outbox_events AS event
      WHERE event.aggregate_type = 'PAYMENT_ORDER'
        AND event.aggregate_id = payment.id
        AND event.event_type = 'booking.appointment.fail'
        AND event.status = 'DELIVERED'
  );

-- Only typed conflict/not-found evidence is strong enough to require a refund.
UPDATE payment_orders AS payment
SET fulfillment_status = CASE
    WHEN EXISTS (
        SELECT 1
        FROM payment_outbox_events AS event
        WHERE event.aggregate_type = 'PAYMENT_ORDER'
          AND event.aggregate_id = payment.id
          AND event.event_type = 'booking.appointment.confirm'
          AND event.status = 'DEAD'
          AND event.terminal_reason_code IN ('conflict', 'not_found')
    ) THEN 'REFUND_REQUIRED'
    ELSE 'MANUAL_REVIEW'
END
WHERE payment.status = 2
  AND payment.fulfillment_status = 'PENDING'
  AND EXISTS (
      SELECT 1
      FROM payment_outbox_events AS event
      WHERE event.aggregate_type = 'PAYMENT_ORDER'
        AND event.aggregate_id = payment.id
        AND event.event_type = 'booking.appointment.confirm'
        AND event.status = 'DEAD'
  );

INSERT INTO payment_compensation_cases (
    payment_order_id,
    appointment_id,
    type,
    status,
    reason_code,
    safe_reason,
    gateway_order_reference,
    gateway_transaction_number,
    gateway_response_code,
    gateway_transaction_status,
    gateway_payment_date,
    amount_vnd,
    created_at,
    updated_at
)
SELECT
    payment.id,
    payment.appointment_id,
    'BOOKING_FULFILLMENT',
    CASE
        WHEN event.terminal_reason_code IN ('conflict', 'not_found') THEN 'REFUND_REQUIRED'
        ELSE 'MANUAL_REVIEW'
    END,
    CASE event.terminal_reason_code
        WHEN 'conflict' THEN 'BOOKING_CONFLICT'
        WHEN 'not_found' THEN 'APPOINTMENT_NOT_FOUND'
        WHEN 'authentication' THEN 'BOOKING_AUTHENTICATION_REJECTED'
        WHEN 'network' THEN 'BOOKING_DELIVERY_RETRIES_EXHAUSTED'
        WHEN 'timeout' THEN 'BOOKING_DELIVERY_RETRIES_EXHAUSTED'
        WHEN 'rate_limited' THEN 'BOOKING_DELIVERY_RETRIES_EXHAUSTED'
        WHEN 'upstream' THEN 'BOOKING_DELIVERY_RETRIES_EXHAUSTED'
        ELSE 'BOOKING_CONTRACT_FAILURE'
    END,
    CASE event.terminal_reason_code
        WHEN 'conflict' THEN 'Booking cannot be confirmed because it is already in an opposite terminal state.'
        WHEN 'not_found' THEN 'Booking appointment was not found after payment succeeded.'
        WHEN 'authentication' THEN 'Booking confirmation was permanently rejected by service authentication.'
        WHEN 'network' THEN 'Booking confirmation exhausted delivery retries.'
        WHEN 'timeout' THEN 'Booking confirmation exhausted delivery retries.'
        WHEN 'rate_limited' THEN 'Booking confirmation exhausted delivery retries.'
        WHEN 'upstream' THEN 'Booking confirmation exhausted delivery retries.'
        ELSE 'Booking confirmation ended with an ambiguous contract or legacy failure.'
    END,
    payment.id::text,
    payment.gateway_txn_ref,
    payment.gateway_response_code,
    payment.gateway_transaction_status,
    payment.gateway_payment_date,
    payment.gross_amount,
    COALESCE(event.last_attempt_at, event.created_at),
    COALESCE(event.last_attempt_at, event.created_at)
FROM payment_orders AS payment
JOIN LATERAL (
    SELECT candidate.*
    FROM payment_outbox_events AS candidate
    WHERE candidate.aggregate_type = 'PAYMENT_ORDER'
      AND candidate.aggregate_id = payment.id
      AND candidate.event_type = 'booking.appointment.confirm'
      AND candidate.status = 'DEAD'
    ORDER BY candidate.created_at DESC, candidate.id DESC
    LIMIT 1
) AS event ON TRUE
WHERE payment.status = 2
  AND payment.appointment_id IS NOT NULL
ON CONFLICT DO NOTHING;

ALTER TABLE payment_orders
    ALTER COLUMN fulfillment_status SET DEFAULT 'PENDING',
    ALTER COLUMN fulfillment_status SET NOT NULL,
    ALTER COLUMN gateway_capture_status SET DEFAULT 'PENDING',
    ALTER COLUMN gateway_capture_status SET NOT NULL;

ALTER TABLE payment_compensation_cases
    ALTER COLUMN gateway_order_reference SET NOT NULL;

ALTER TABLE payment_orders
    DROP CONSTRAINT IF EXISTS chk_payment_orders_fulfillment_status;
ALTER TABLE payment_orders
    ADD CONSTRAINT chk_payment_orders_fulfillment_status
    CHECK (fulfillment_status IN ('PENDING', 'BOOKING_CONFIRMED', 'BOOKING_FAILED', 'MANUAL_REVIEW', 'REFUND_REQUIRED'));

ALTER TABLE payment_orders
    DROP CONSTRAINT IF EXISTS chk_payment_orders_gateway_capture_status;
ALTER TABLE payment_orders
    ADD CONSTRAINT chk_payment_orders_gateway_capture_status
    CHECK (gateway_capture_status IN ('PENDING', 'CAPTURED', 'FAILED', 'CAPTURED_DUPLICATE'));

ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_type;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_type
    CHECK (type IN ('BOOKING_FULFILLMENT', 'DUPLICATE_GATEWAY_CAPTURE'));

ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_status;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_status
    CHECK (status IN ('MANUAL_REVIEW', 'REFUND_REQUIRED'));

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
        'DUPLICATE_GATEWAY_CAPTURE'
    ));

ALTER TABLE payment_compensation_cases
    DROP CONSTRAINT IF EXISTS chk_payment_compensation_cases_amount;
ALTER TABLE payment_compensation_cases
    ADD CONSTRAINT chk_payment_compensation_cases_amount
    CHECK (amount_vnd > 0);

CREATE INDEX IF NOT EXISTS ix_payment_compensation_cases_status_created
    ON payment_compensation_cases (status, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_payment_compensation_cases_appointment
    ON payment_compensation_cases (appointment_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_payment_compensation_cases_payment_order
    ON payment_compensation_cases (payment_order_id, created_at DESC);

COMMIT;
