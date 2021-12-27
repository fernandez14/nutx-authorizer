package authorizer

import "bytes"

// Violation represents transaction violations.
type Violation int

const (
	AccountAlreadyInitialized Violation = iota
	AccountNotInitialized
	CardNotActive
	InsufficientLimit
	HighFrequencySmallInterval
	DoubledTransaction
)

func (v Violation) String() string {
	switch v {
	case AccountAlreadyInitialized:
		return "account-already-initialized"
	case AccountNotInitialized:
		return "account-not-initialized"
	case CardNotActive:
		return "card-not-active"
	case InsufficientLimit:
		return "insufficient-limit"
	case HighFrequencySmallInterval:
		return "high-frequency-small-interval"
	case DoubledTransaction:
		return "doubled-transaction"
	}
	panic("invalid violation enum option")
}

func (v Violation) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(v.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

type Violations []Violation

func (v Violations) Has(item Violation) bool {
	for _, current := range v {
		if current == item {
			return true
		}
	}
	return false
}
