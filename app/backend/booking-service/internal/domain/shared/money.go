package shared

import (
	"errors"
	"fmt"
	"math"
)

var ErrInvalidMoneyVND = errors.New("invalid VND money")

type MoneyVND int64

func NewMoneyVNDFromPrice(price float64) (MoneyVND, error) {
	if math.IsNaN(price) {
		return 0, fmt.Errorf("%w: price is NaN", ErrInvalidMoneyVND)
	}
	if math.IsInf(price, 0) {
		return 0, fmt.Errorf("%w: price is infinite", ErrInvalidMoneyVND)
	}
	if price <= 0 {
		return 0, fmt.Errorf("%w: price must be greater than zero", ErrInvalidMoneyVND)
	}
	if math.Trunc(price) != price {
		return 0, fmt.Errorf("%w: fractional VND is not supported", ErrInvalidMoneyVND)
	}
	const maxVNPaySafeAmountVND = float64(92233720368547758)
	if price >= maxVNPaySafeAmountVND {
		return 0, fmt.Errorf("%w: price is too large", ErrInvalidMoneyVND)
	}
	return MoneyVND(price), nil
}
