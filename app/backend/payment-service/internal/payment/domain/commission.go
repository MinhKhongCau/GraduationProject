package domain

import "payment-service/internal/domain/vo"

const DefaultCommissionRate = 0.15

type Commission struct {
	Rate      float64
	Amount    vo.Money
	NetAmount vo.Money
}

func CalculateCommission(grossAmount vo.Money) Commission {
	commissionAmount := vo.Money(float64(grossAmount) * DefaultCommissionRate)
	netAmount := grossAmount.Sub(commissionAmount)
	return Commission{
		Rate:      DefaultCommissionRate,
		Amount:    commissionAmount,
		NetAmount: netAmount,
	}
}
