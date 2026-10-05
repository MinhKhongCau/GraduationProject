package payment

import "payment-service/internal/domain/money"

const DefaultCommissionRate = 0.15

type Commission struct {
	Rate      float64
	Amount    money.Money
	NetAmount money.Money
}

func CalculateCommission(grossAmount money.Money) Commission {
	commissionAmount := money.Money(float64(grossAmount) * DefaultCommissionRate)
	netAmount := grossAmount.Sub(commissionAmount)
	return Commission{
		Rate:      DefaultCommissionRate,
		Amount:    commissionAmount,
		NetAmount: netAmount,
	}
}
