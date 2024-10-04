// Package bigrat provides high-precision numerical operations.
package bigrat

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"

	"hw/pkg/logger"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

// BigN represents a high-precision number and its possible error.
type BigN struct {
	num decimal.Decimal
	err error
	mu  sync.Mutex
}

// NewBigN creates a new instance of BigN.
func NewBigN(num interface{}) *BigN {
	dec, err := coverToDecimal(num)
	bn := &BigN{num: dec, err: nil}
	if err != nil {
		bn.err = err
		return bn
	}
	return bn
}

// Add adds the given number to BigN.
func (bn *BigN) Add(n interface{}) *BigN {
	newBN := &BigN{}

	bn.mu.Lock()
	defer bn.mu.Unlock()

	if bn.err != nil {
		newBN.err = bn.err
		return newBN
	}

	dec, err := coverToDecimal(n)
	if err != nil {
		newBN.err = err
		return newBN
	}

	newBN.num = bn.num.Add(dec)
	return newBN
}

// Sub subtracts the given number from BigN.
func (bn *BigN) Sub(n interface{}) *BigN {
	newBN := &BigN{}

	bn.mu.Lock()
	defer bn.mu.Unlock()

	if bn.err != nil {
		newBN.err = bn.err
		return newBN
	}

	dec, err := coverToDecimal(n)
	if err != nil {
		newBN.err = err
		return newBN
	}

	newBN.num = bn.num.Sub(dec)
	return newBN
}

// Mul multiplies BigN by the given number.
func (bn *BigN) Mul(n interface{}) *BigN {
	newBN := &BigN{}

	bn.mu.Lock()
	defer bn.mu.Unlock()

	if bn.err != nil {
		newBN.err = bn.err
		return newBN
	}

	dec, err := coverToDecimal(n)
	if err != nil {
		newBN.err = err
		return newBN
	}

	newBN.num = bn.num.Mul(dec)
	return newBN
}

// Pow raises BigN to the given exponent.
func (bn *BigN) Pow(n int64) *BigN {
	newBN := &BigN{}

	bn.mu.Lock()
	defer bn.mu.Unlock()

	if n < 0 {
		newBN.err = fmt.Errorf("invalid exponent: negative value")
		return newBN
	}

	exponent := decimal.NewFromInt(n)
	newBN.num = bn.num.Pow(exponent)
	return newBN
}

// Div divides BigN by the given number.
func (bn *BigN) Div(n interface{}) *BigN {
	newBN := &BigN{}

	bn.mu.Lock()
	defer bn.mu.Unlock()

	if bn.err != nil {
		newBN.err = bn.err
		return newBN
	}

	d, err := coverToDecimal(n)
	if err != nil {
		newBN.err = err
		return newBN
	}

	if d.IsZero() {
		pc, file, line, ok := runtime.Caller(1)
		if !ok {
			newBN.err = fmt.Errorf("no caller information")
			return newBN
		}
		fn := runtime.FuncForPC(pc)
		newBN.err = fmt.Errorf("division by zero at %s - %s:%d", fn.Name(), file, line)
		logger.Warnf("b.num %+v, div num is zero %+v %+v", bn.num.String(), n, newBN.err)
		return newBN
	}

	newBN.num = bn.num.Div(d)
	return newBN
}

// ToTruncateString truncates BigN to the specified number of decimal places and returns it as a string.
func (bn *BigN) ToTruncateString(d int32) string {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	if d < 0 {
		bn.err = fmt.Errorf("invalid decimal places: negative value")
		return bn.num.String()
	}
	return bn.num.Truncate(d).StringFixed(d)
}

// ToTruncateInt64 truncates BigN to the specified number of decimal places and returns it as int64.
func (bn *BigN) ToTruncateInt64(d int32) int64 {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	if d < 0 {
		bn.err = fmt.Errorf("invalid decimal places: negative value")
		return 0
	}
	return bn.num.Truncate(d).IntPart()
}

// ToTruncateFloat64 truncates BigN to the specified number of decimal places and returns it as float64.
func (bn *BigN) ToTruncateFloat64(d int32) float64 {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	if d < 0 {
		bn.err = fmt.Errorf("invalid decimal places: negative value")
		return 0.0
	}
	f64, exact := bn.num.Truncate(d).Float64()
	if !exact {
		logger.Warnf("bn.num %+v, truncate to float64 %+v, not exact", bn.num.String(), f64)
	}
	return f64
}

// Error returns the error in BigN.
func (bn *BigN) Error() error {
	bn.mu.Lock()
	defer bn.mu.Unlock()
	return bn.err
}

// ToMoneyString formats BigN as a currency string with thousand separators.
func (bn *BigN) ToMoneyString(decimals int32) string {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	// Truncate the number to the specified decimal places
	truncated := bn.num.Truncate(decimals)

	// Convert the truncated number to string with fixed decimal places
	str := truncated.StringFixed(decimals)

	// Split the string into integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = parts[1]
	}

	// Add thousand separators to the integer part
	var result strings.Builder
	if intPart[0] == '-' {
		result.WriteByte('-')
		intPart = intPart[1:]
	}
	for i, digit := range intPart {
		if i != 0 && (len(intPart)-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(digit)
	}

	// If there is a decimal part, append it to the formatted integer part
	if decimals > 0 {
		result.WriteByte('.')
		// Ensure the decimal part has the correct length by padding with zeros if necessary
		decPart = fmt.Sprintf("%-*s", decimals, decPart)
		decPart = strings.Replace(decPart, " ", "0", -1)
		result.WriteString(decPart)
	}

	return result.String()
}

// coverToDecimal converts different types of numbers to decimal.Decimal.
func coverToDecimal(num interface{}) (decimal.Decimal, error) {
	switch v := num.(type) {
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case uint8:
		return decimal.NewFromInt(cast.ToInt64(v)), nil
	case uint64:
		bigInt := new(big.Int).SetUint64(v)
		return decimal.NewFromBigInt(bigInt, 0), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case string:
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
			v = v[2:]
		}
		if dec, err := decimal.NewFromString(v); err == nil {
			return dec, nil
		}
		if bigInt, success := new(big.Int).SetString(v, 16); success {
			return decimal.NewFromBigInt(bigInt, 0), nil
		}
		return decimal.Zero, fmt.Errorf("coverToDecimal: unsupported string format")
	case *big.Int:
		return decimal.NewFromBigInt(v, 0), nil
	case *BigN:
		return v.num, nil
	default:
		return decimal.Zero, fmt.Errorf("coverToDecimal: unsupported type %T", num)
	}
}
