package authorizer

import (
	"encoding/json"
	"time"
)

type Account struct {
	Active         bool `json:"active-card"`
	AvailableLimit int  `json:"available-limit"`
}

func (acc Account) MarshalJSON() ([]byte, error) {
	// we need a local alias type to break the json.Marshal loop.
	type aliased Account
	empty := Account{}
	if acc == empty {
		return []byte("{}"), nil
	}
	return json.Marshal(aliased(acc))
}

type Transaction struct {
	Merchant string    `json:"merchant"`
	Amount   int       `json:"amount"`
	Time     time.Time `json:"time"`
}

type Operation struct {
	Account     *Account     `json:"account"`
	Transaction *Transaction `json:"transaction"`
}

func (op Operation) toSuccessOutput() OperationOutput {
	out := OperationOutput{
		Input:      op,
		Violations: Violations{},
	}
	if op.Account != nil {
		out.Account = *op.Account
	}
	return out
}

type OperationOutput struct {
	At         time.Time  `json:"-"`
	Input      Operation  `json:"-"`
	Account    Account    `json:"account"`
	Violations Violations `json:"violations"`
}

type OperationOutputList []OperationOutput

func (list OperationOutputList) IsDoubledTransactionViolated(op Operation, t time.Time) bool {
	if op.Transaction == nil {
		return false
	}
	var count int
	bound := t.Add(-2 * time.Minute)
	// since list is in ascending order we iterate in reverse to stop early.
	for n := len(list) - 1; n >= 0; n-- {
		out := list[n]
		if out.Input.Transaction == nil {
			continue
		}
		at := out.Input.Transaction.Time
		if !at.After(bound) {
			break
		}
		if out.Input.Transaction.Merchant == op.Transaction.Merchant && out.Input.Transaction.Amount == op.Transaction.Amount {
			count++
		}
	}
	return count >= 1
}

func (list OperationOutputList) IsHighFrequencyViolated(t time.Time) bool {
	var count int
	bound := t.Add(-2 * time.Minute)
	// since list is in ascending order we iterate in reverse to stop early.
	for n := len(list) - 1; n >= 0; n-- {
		op := list[n]
		if op.Input.Transaction == nil {
			continue
		}
		at := op.Input.Transaction.Time
		if !at.After(bound) {
			break
		}
		count++
	}
	return count >= 3
}
