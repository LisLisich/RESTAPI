package domain

import (
	"fmt"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type Currency string

const CurrencyRUB Currency = "RUB"

type Money struct {
	MinorUnits int64
	Currency   Currency
}

func NewMoney(minorUnits int64, currency Currency) (Money, error) {
	if currency != CurrencyRUB {
		return Money{}, fmt.Errorf(
			"unsupported currency %q: %w",
			currency,
			core_errors.ErrInvalidArgument,
		)
	}
	return Money{MinorUnits: minorUnits, Currency: currency}, nil
}

func NewPaymentAmount(minorUnits int64) (Money, error) {
	if minorUnits <= 0 {
		return Money{}, fmt.Errorf(
			"payment amount must be positive: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	return NewMoney(minorUnits, CurrencyRUB)
}
