package authorizer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func runGopher(ops chan Operation) chan OperationOutput {
	results := make(chan OperationOutput)
	go func() {
		var (
			account *Account
			settled OperationOutputList
		)
		settle := func(op OperationOutput) {
			op.At = time.Now()
			// Authorizer operations that had violations should not be saved in the application's internal state.
			if len(op.Violations) == 0 {
				settled = append(settled, op)
				//settled = settled.Purge(op)
			}
			results <- op
		}
		for op := range ops {
			// we treat the operation struct as an enum.
			switch true {
			case op.Account != nil:
				// check for initialization errors.
				if account != nil {
					settle(OperationOutput{
						Input:      op,
						Account:    *account,
						Violations: Violations{AccountAlreadyInitialized},
					})
					continue
				}
				account = op.Account
				settle(op.toSuccessOutput())
			case op.Transaction != nil:
				out := OperationOutput{
					Input:      op,
					Violations: Violations{},
				}
				if account == nil {
					out.Violations = append(out.Violations, AccountNotInitialized)
					settle(out)
					continue
				}
				out.Account = *account
				if account.Active == false {
					out.Violations = append(out.Violations, CardNotActive)
				}
				if account.AvailableLimit < op.Transaction.Amount {
					out.Violations = append(out.Violations, InsufficientLimit)
				}
				if settled.IsHighFrequencyViolated(op.Transaction.Time) {
					out.Violations = append(out.Violations, HighFrequencySmallInterval)
				}
				if settled.IsDoubledTransactionViolated(op, op.Transaction.Time) {
					out.Violations = append(out.Violations, DoubledTransaction)
				}
				if len(out.Violations) == 0 {
					account.AvailableLimit -= op.Transaction.Amount
					out.Account = *account
				}
				settle(out)
			default:
				panic("invalid operation")
			}
		}
		close(results)
	}()
	return results
}

// Scanner ingests data from an io.Reader and sends it to an authorizer worker for tx processing. It outputs results to an io.Writer.
func Scanner(in io.Reader, out io.Writer) {
	ops := make(chan Operation)
	transactor := runGopher(ops)
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		var op Operation
		err := json.Unmarshal(scanner.Bytes(), &op)
		if err != nil {
			out.Write([]byte(fmt.Sprintf("json.Unmarshal error: %v", err)))
			continue
		}
		ops <- op
		res := <-transactor
		b, err := json.Marshal(res)
		if err != nil {
			out.Write([]byte(fmt.Sprintf("json.Marshal error: %v", err)))
		}
		out.Write(b)
		out.Write([]byte("\n"))
	}
}
