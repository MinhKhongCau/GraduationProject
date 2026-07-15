package vo

type Money int64

func (m Money) Int64() int64 {
	return int64(m)
}

func (m Money) Add(other Money) Money {
	return m + other
}

func (m Money) Sub(other Money) Money {
	return m - other
}

func (m Money) MulFloat(factor float64) Money {
	return Money(float64(m) * factor)
}

func (m Money) IsZero() bool {
	return m == 0
}

func (m Money) IsPositive() bool {
	return m > 0
}

func (m Money) IsNegative() bool {
	return m < 0
}
