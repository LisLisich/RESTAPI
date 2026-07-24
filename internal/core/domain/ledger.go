package domain

import (
	"fmt"
	"math/big"
	"strings"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type LedgerPosting struct {
	Account string
	Money   Money
}

type LedgerTransaction struct {
	ID       string
	Postings []LedgerPosting
}

func NewLedgerTransaction(
	id string,
	postings []LedgerPosting,
) (LedgerTransaction, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return LedgerTransaction{}, fmt.Errorf(
			"ledger transaction id is required: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if len(postings) < 2 {
		return LedgerTransaction{}, fmt.Errorf(
			"ledger transaction requires at least two postings: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	sum := new(big.Int)
	for index, posting := range postings {
		if strings.TrimSpace(posting.Account) == "" {
			return LedgerTransaction{}, fmt.Errorf(
				"ledger posting %d account is required: %w",
				index,
				core_errors.ErrInvalidArgument,
			)
		}
		if posting.Money.Currency != CurrencyRUB {
			return LedgerTransaction{}, fmt.Errorf(
				"ledger posting %d has unsupported currency %q: %w",
				index,
				posting.Money.Currency,
				core_errors.ErrInvalidArgument,
			)
		}
		if posting.Money.MinorUnits == 0 {
			return LedgerTransaction{}, fmt.Errorf(
				"ledger posting %d amount must be non-zero: %w",
				index,
				core_errors.ErrInvalidArgument,
			)
		}
		sum.Add(sum, big.NewInt(posting.Money.MinorUnits))
	}
	if sum.Sign() != 0 {
		return LedgerTransaction{}, fmt.Errorf(
			"ledger postings are not balanced, delta is %s: %w",
			sum.String(),
			core_errors.ErrInvalidArgument,
		)
	}

	return LedgerTransaction{
		ID:       id,
		Postings: append([]LedgerPosting(nil), postings...),
	}, nil
}
