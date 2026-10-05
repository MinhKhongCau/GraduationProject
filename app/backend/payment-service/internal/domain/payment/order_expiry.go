package payment

import "time"

func PaymentExpiry(now time.Time, ttl time.Duration, bookingExpiresAt int64) int64 {
	ttlExpiry := now.Add(ttl).UnixMilli()
	if bookingExpiresAt < ttlExpiry {
		ttlExpiry = bookingExpiresAt
	}
	return time.UnixMilli(ttlExpiry).Truncate(time.Second).UnixMilli()
}

func HasUsablePaymentWindow(now time.Time, expiresAt int64, minimumWindow time.Duration) bool {
	return expiresAt-now.UnixMilli() >= minimumWindow.Milliseconds()
}

func IsOrderExpired(now time.Time, expiresAt int64) bool {
	return now.UnixMilli() >= expiresAt
}
