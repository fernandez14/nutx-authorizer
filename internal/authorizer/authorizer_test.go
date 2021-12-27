package authorizer

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input: `{"account": {"active-card": true, "available-limit": 100}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T10:00:00.000Z"}}
{"transaction": {"merchant": "Habbib's", "amount": 90, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "McDonald's", "amount": 30, "time": "2019-02-13T12:00:00.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":100},"violations":[]}
{"account":{"active-card":true,"available-limit":80},"violations":[]}
{"account":{"active-card":true,"available-limit":80},"violations":["insufficient-limit"]}
{"account":{"active-card":true,"available-limit":50},"violations":[]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 175}}
{"account": {"active-card": true, "available-limit": 350}}`,
			output: `{"account":{"active-card":true,"available-limit":175},"violations":[]}
{"account":{"active-card":true,"available-limit":175},"violations":["account-already-initialized"]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 100}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T11:00:00.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":100},"violations":[]}
{"account":{"active-card":true,"available-limit":80},"violations":[]}`,
		},
		{
			input: `{"transaction": {"merchant": "Uber Eats", "amount": 25, "time": "2020-12-01T11:07:00.000Z"}}
{"account": {"active-card": true, "available-limit": 225}}
{"transaction": {"merchant": "Uber Eats", "amount": 25, "time": "2020-12-01T11:07:00.000Z"}}`,
			output: `{"account":{},"violations":["account-not-initialized"]}
{"account":{"active-card":true,"available-limit":225},"violations":[]}
{"account":{"active-card":true,"available-limit":200},"violations":[]}`,
		},
		{
			input: `{"account": {"active-card": false, "available-limit": 100}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "Habbib's", "amount": 15, "time": "2019-02-13T11:15:00.000Z"}}`,
			output: `{"account":{"active-card":false,"available-limit":100},"violations":[]}
{"account":{"active-card":false,"available-limit":100},"violations":["card-not-active"]}
{"account":{"active-card":false,"available-limit":100},"violations":["card-not-active"]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 1000}}
{"transaction": {"merchant": "Vivara", "amount": 1250, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "Samsung", "amount": 2500, "time": "2019-02-13T11:00:01.000Z"}}
{"transaction": {"merchant": "Nike", "amount": 800, "time": "2019-02-13T11:01:01.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":1000},"violations":[]}
{"account":{"active-card":true,"available-limit":1000},"violations":["insufficient-limit"]}
{"account":{"active-card":true,"available-limit":1000},"violations":["insufficient-limit"]}
{"account":{"active-card":true,"available-limit":200},"violations":[]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 100}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "Habbib's", "amount": 20, "time": "2019-02-13T11:00:01.000Z"}}
{"transaction": {"merchant": "McDonald's", "amount": 20, "time": "2019-02-13T11:01:01.000Z"}}
{"transaction": {"merchant": "Subway", "amount": 20, "time": "2019-02-13T11:01:31.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 10, "time": "2019-02-13T12:00:00.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":100},"violations":[]}
{"account":{"active-card":true,"available-limit":80},"violations":[]}
{"account":{"active-card":true,"available-limit":60},"violations":[]}
{"account":{"active-card":true,"available-limit":40},"violations":[]}
{"account":{"active-card":true,"available-limit":40},"violations":["high-frequency-small-interval"]}
{"account":{"active-card":true,"available-limit":30},"violations":[]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 100}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "McDonald's", "amount": 10, "time": "2019-02-13T11:00:01.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T11:00:02.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 15, "time": "2019-02-13T11:00:03.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":100},"violations":[]}
{"account":{"active-card":true,"available-limit":80},"violations":[]}
{"account":{"active-card":true,"available-limit":70},"violations":[]}
{"account":{"active-card":true,"available-limit":70},"violations":["doubled-transaction"]}
{"account":{"active-card":true,"available-limit":55},"violations":[]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 100}}
{"transaction": {"merchant": "McDonald's", "amount": 10, "time": "2019-02-13T11:00:01.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T11:00:02.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 5, "time": "2019-02-13T11:00:07.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 5, "time": "2019-02-13T11:00:08.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 150, "time": "2019-02-13T11:00:18.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 190, "time": "2019-02-13T11:00:22.000Z"}}
{"transaction": {"merchant": "Burger King", "amount": 15, "time": "2019-02-13T12:00:27.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":100},"violations":[]}
{"account":{"active-card":true,"available-limit":90},"violations":[]}
{"account":{"active-card":true,"available-limit":70},"violations":[]}
{"account":{"active-card":true,"available-limit":65},"violations":[]}
{"account":{"active-card":true,"available-limit":65},"violations":["high-frequency-small-interval","doubled-transaction"]}
{"account":{"active-card":true,"available-limit":65},"violations":["insufficient-limit","high-frequency-small-interval"]}
{"account":{"active-card":true,"available-limit":65},"violations":["insufficient-limit","high-frequency-small-interval"]}
{"account":{"active-card":true,"available-limit":50},"violations":[]}`,
		},
		{
			input: `{"account": {"active-card": true, "available-limit": 1000}}
{"transaction": {"merchant": "Vivara", "amount": 1250, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "Samsung", "amount": 2500, "time": "2019-02-13T11:00:01.000Z"}}
{"transaction": {"merchant": "Nike", "amount": 800, "time": "2019-02-13T11:01:01.000Z"}}
{"transaction": {"merchant": "Uber", "amount": 80, "time": "2019-02-13T11:01:31.000Z"}}`,
			output: `{"account":{"active-card":true,"available-limit":1000},"violations":[]}
{"account":{"active-card":true,"available-limit":1000},"violations":["insufficient-limit"]}
{"account":{"active-card":true,"available-limit":1000},"violations":["insufficient-limit"]}
{"account":{"active-card":true,"available-limit":200},"violations":[]}
{"account":{"active-card":true,"available-limit":120},"violations":[]}`,
		},
	}
	for _, test := range tests {
		out := new(bytes.Buffer)
		Scanner(bytes.NewBufferString(test.input), out)
		assert.Equal(t, test.output+"\n", out.String())
	}
}

func benchmarkScanner(b *testing.B, lines int) {
	var buff bytes.Buffer
	buff.WriteString(`{"account": {"active-card": true, "available-limit": 1000000}}`)
	buff.WriteString("\n")
	now := time.Now()
	for i := 0; i < lines; i++ {
		buff.WriteString(`{"transaction": {"merchant": "Vivara", "amount": 100, "time": "`+now.Add(time.Duration(i)*time.Minute).String()+`"}}`)
		buff.WriteString("\n")
	}
	data := buff.Bytes()
	for i := 0; i < b.N; i++ {
		out := new(bytes.Buffer)
		Scanner(bytes.NewBuffer(data), out)
		assert.Greater(b, out.Len(), 0)
	}
}

func BenchmarkScanner10(b *testing.B) {
	benchmarkScanner(b, 10)
}

func BenchmarkScanner100(b *testing.B) {
	benchmarkScanner(b, 100)
}

func BenchmarkScanner1000(b *testing.B) {
	benchmarkScanner(b, 1000)
}

func BenchmarkScanner10000(b *testing.B) {
	benchmarkScanner(b, 10000)
}


