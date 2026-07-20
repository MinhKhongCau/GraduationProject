package domain

import (
	"errors"
	"fmt"
	"strconv"
)

func ParseVNPayAmount(params map[string][]string) (int64, error) {
	vnpAmountStr := ""
	if len(params["vnp_Amount"]) > 0 {
		vnpAmountStr = params["vnp_Amount"][0]
	}
	if vnpAmountStr == "" {
		return 0, errors.New("missing vnp_Amount")
	}
	vnpAmount, err := strconv.ParseInt(vnpAmountStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid vnp_Amount: %w", err)
	}
	if vnpAmount <= 0 {
		return 0, errors.New("invalid vnp_Amount: amount must be greater than zero")
	}
	return vnpAmount / 100, nil
}
