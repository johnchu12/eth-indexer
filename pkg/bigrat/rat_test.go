package bigrat

import (
	"math/big"
	"testing"
)

func TestCustomizeBigNOperations(t *testing.T) {
	type testCase struct {
		input       func() interface{}
		expected    interface{}
		description string
	}

	testCases := []testCase{
		{func() interface{} { return NewBigN("1.23456").ToTruncateString(1) }, "1.2", "1.23456 truncate to 1 decimal"},
		{func() interface{} { return NewBigN("1.1").ToTruncateString(1) }, "1.1", "1.1 truncate to 1 decimal"},
		{func() interface{} { return NewBigN("1.55456").ToTruncateString(1) }, "1.5", "1.55456 truncate to 1 decimal"},
		{func() interface{} { return NewBigN("1.23454e-06").ToTruncateString(10) }, "0.0000012345", "1.23454e-06 truncate to 10 decimals"},
		{func() interface{} { return NewBigN("1.23456e-06").ToTruncateString(10) }, "0.0000012345", "1.23456e-06 truncate to 10 decimals"},
		{func() interface{} { return NewBigN("1.23456e9").ToTruncateString(0) }, "1234560000", "1.23456e9 truncate to 0 decimals"},
		{func() interface{} { return NewBigN(1).Add("0xf").ToTruncateString(0) }, "16", "1 + 0xF"},
		{func() interface{} { return NewBigN("0xF").ToTruncateString(0) }, "15", "0xF to decimal"},
		{func() interface{} { return NewBigN("0").Add("3501458324437618").ToTruncateString(0) }, "3501458324437618", "0 + 3501458324437618"},
		{func() interface{} { return NewBigN("0xC70D815D562D3CFA955").ToTruncateString(0) }, "58750003716598352816469", "0xC70D815D562D3CFA955 to decimal"},
		{func() interface{} { return NewBigN(2).Pow(256).ToTruncateString(0) }, "115792089237316195423570985008687907853269984665640564039457584007913129639936", "2^256"},
		{func() interface{} { return NewBigN(1.23456e9).ToTruncateString(0) }, "1234560000", "1.23456e9 truncate to 0 decimals"},
		{func() interface{} { return NewBigN(1.23456).ToTruncateString(2) }, "1.23", "1.23456 truncate to 2 decimals"},
		{func() interface{} { return NewBigN(1.23956).ToTruncateString(2) }, "1.23", "1.23956 truncate to 2 decimals"},
		{func() interface{} { return NewBigN(0.23456).ToTruncateString(2) }, "0.23", "0.23456 truncate to 2 decimals"},
		{func() interface{} { return NewBigN(0.23956).ToTruncateString(2) }, "0.23", "0.23956 truncate to 2 decimals"},
		{func() interface{} { return NewBigN(38952000000000).ToTruncateString(0) }, "38952000000000", "38952000000000 truncate to 0 decimals"},
		{func() interface{} { return NewBigN(1.23456).ToTruncateFloat64(1) }, 1.2, "1.23456 truncate to 1 decimal (float64)"},
		{func() interface{} { return NewBigN(1.1).ToTruncateFloat64(1) }, 1.1, "1.1 truncate to 1 decimal (float64)"},
		{func() interface{} { return NewBigN(1.55456).ToTruncateFloat64(1) }, 1.5, "1.55456 truncate to 1 decimal (float64)"},
		{func() interface{} { return NewBigN("1.23454e-06").ToTruncateFloat64(10) }, 0.0000012345, "1.23454e-06 truncate to 10 decimals (float64)"},
		{func() interface{} { return NewBigN("1.23456e-06").ToTruncateFloat64(10) }, 0.0000012345, "1.23456e-06 truncate to 10 decimals (float64)"},
		{func() interface{} { return NewBigN("1.23456e9").ToTruncateFloat64(0) }, 1234560000.0, "1.23456e9 truncate to 0 decimals (float64)"},
		{func() interface{} { return NewBigN(1).Add("0xf").ToTruncateFloat64(0) }, 16.0, "1 + 0xF (float64)"},
		{func() interface{} { return NewBigN(15).Add("0xf").ToTruncateFloat64(0) }, 30.0, "15 + 0xF (float64)"},
		{func() interface{} { return NewBigN("0xF").ToTruncateFloat64(0) }, 15.0, "0xF to decimal (float64)"},
		{func() interface{} { return NewBigN("0xC70D815D562D3CFA955").ToTruncateFloat64(0) }, 58750003716598352816469.0, "0xC70D815D562D3CFA955 to decimal (float64)"},
		{func() interface{} { return NewBigN(1.23456e9).ToTruncateFloat64(0) }, 1234560000.0, "1.23456e9 truncate to 0 decimals (float64)"},
		{func() interface{} { return NewBigN(1.23456).ToTruncateFloat64(2) }, 1.23, "1.23456 truncate to 2 decimals (float64)"},
		{func() interface{} { return NewBigN(1.23956).ToTruncateFloat64(2) }, 1.23, "1.23956 truncate to 2 decimals (float64)"},
		{func() interface{} { return NewBigN(0.23456).ToTruncateFloat64(2) }, 0.23, "0.23456 truncate to 2 decimals (float64)"},
		{func() interface{} { return NewBigN(0.23956).ToTruncateFloat64(2) }, 0.23, "0.23956 truncate to 2 decimals (float64)"},
		{func() interface{} { return NewBigN(38952000000000).ToTruncateFloat64(0) }, 38952000000000.0, "38952000000000 truncate to 0 decimals (float64)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.input()

			switch expected := tc.expected.(type) {
			case string:
				if resultStr, ok := result.(string); !ok {
					t.Errorf("%s fail: got type %T, want string", tc.description, result)
				} else if resultStr != expected {
					t.Errorf("%s fail: got %v, want %v", tc.description, resultStr, expected)
				}
			case float64:
				if resultFloat, ok := result.(float64); !ok {
					t.Errorf("%s fail: got type %T, want float64", tc.description, result)
				} else if resultFloat != expected {
					t.Errorf("%s fail: got %v, want %v", tc.description, resultFloat, expected)
				}
			case int64:
				if resultInt, ok := result.(int64); !ok {
					t.Errorf("%s failed: got type %T, want int64", tc.description, result)
				} else if resultInt != expected {
					t.Errorf("%s failed: got %v, want %v", tc.description, resultInt, expected)
				}
			default:
				t.Fatalf("Unsupported expected type: %T", expected)
			}
		})
	}
}

// TestCoverToDecimal tests the coverToDecimal function.
func TestCoverToDecimal(t *testing.T) {
	type testCase struct {
		input       interface{}
		expected    string
		expectError bool
		description string
	}

	testCases := []testCase{
		// Valid input types
		{input: 42, expected: "42", expectError: false, description: "int to decimal"},
		{input: int64(42), expected: "42", expectError: false, description: "int64 to decimal"},
		{input: uint8(255), expected: "255", expectError: false, description: "uint8 to decimal"},
		{input: uint64(4294967295), expected: "4294967295", expectError: false, description: "uint64 to decimal"},
		{input: 3.14159, expected: "3.14159", expectError: false, description: "float64 to decimal"},
		{input: "123.456", expected: "123.456", expectError: false, description: "string decimal to decimal"},
		{input: "0xFF", expected: "255", expectError: false, description: "hex string to decimal"},
		{input: new(big.Int).SetUint64(1000), expected: "1000", expectError: false, description: "*big.Int to decimal"},
		{input: NewBigN("500.5"), expected: "500.5", expectError: false, description: "*BigN to decimal"},

		// Edge cases and invalid inputs
		{input: "invalid", expected: "0", expectError: true, description: "invalid string"},
		{input: nil, expected: "0", expectError: true, description: "nil input"},
		{input: struct{}{}, expected: "0", expectError: true, description: "unsupported type"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dec, err := coverToDecimal(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for input %v, but got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error for input %v, but got %v", tc.input, err)
				}
				if dec.String() != tc.expected {
					t.Errorf("Expected %v, but got %v", tc.expected, dec.String())
				}
			}
		})
	}
}

func TestSubOperations(t *testing.T) {
	testCases := []struct {
		input1      interface{}
		input2      interface{}
		expected    string
		description string
	}{
		{NewBigN("10.5"), "5.2", "5.3000", "10.5 - 5.2 = 5.3000"},
		{NewBigN("0"), "0", "0.0000", "0 - 0 = 0.0000"},
		{NewBigN("-5"), "3", "-8.0000", "-5 - 3 = -8.0000"},
		{NewBigN("1.23456e3"), "1.23456e2", "1111.1040", "1.23456e3 - 1.23456e2 = 1111.1040"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.input1.(*BigN).Sub(tc.input2).ToTruncateString(4)
			if result != tc.expected {
				t.Errorf("Sub operation failed: got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestMulOperations(t *testing.T) {
	testCases := []struct {
		input1      interface{}
		input2      interface{}
		expected    string
		description string
	}{
		{NewBigN("2"), "3", "6.0000", "2 * 3 = 6.0000"},
		{NewBigN("-1.5"), "4", "-6.0000", "-1.5 * 4 = -6.0000"},
		{NewBigN("1.234"), "0", "0.0000", "1.234 * 0 = 0.0000"},
		{NewBigN("1.2e3"), "2.5e2", "300000.0000", "1.2e3 * 2.5e2 = 300000.0000"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.input1.(*BigN).Mul(tc.input2).ToTruncateString(4)
			if result != tc.expected {
				t.Errorf("Mul operation failed: got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestDivOperations(t *testing.T) {
	testCases := []struct {
		input1      interface{}
		input2      interface{}
		expected    string
		description string
	}{
		{NewBigN("10"), "2", "5.0000", "10 / 2 = 5.0000"},
		{NewBigN("7.5"), "2.5", "3.0000", "7.5 / 2.5 = 3.0000"},
		{NewBigN("1"), "3", "0.3333", "1 / 3 = 0.3333"},
		{NewBigN("1.23456e3"), "1.23456e2", "10.0000", "1.23456e3 / 1.23456e2 = 10.0000"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.input1.(*BigN).Div(tc.input2).ToTruncateString(4)
			if result != tc.expected {
				t.Errorf("Div operation failed: got %v, want %v", result, tc.expected)
			}
		})
	}

	t.Run("division by zero", func(t *testing.T) {
		bn := NewBigN("10")
		result := bn.Div("0")
		if result.Error() == nil {
			t.Errorf("Expected error for division by zero, got nil")
		}
	})
}

func TestToTruncateInt64(t *testing.T) {
	testCases := []struct {
		input       interface{}
		decimals    int32
		expected    int64
		description string
	}{
		{NewBigN("123.456"), 0, 123, "Truncate 123.456 to 0 decimals"},
		{NewBigN("123.456"), 2, 123, "Truncate 123.456 to 2 decimals"},
		{NewBigN("-789.654"), 1, -789, "Truncate -789.654 to 1 decimal"},
		{NewBigN("0.999"), 0, 0, "Truncate 0.999 to 0 decimals"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.input.(*BigN).ToTruncateInt64(tc.decimals)
			if result != tc.expected {
				t.Errorf("ToTruncateInt64 failed: got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestErrorMethod(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		bn := NewBigN("100")
		if bn.Error() != nil {
			t.Errorf("Expected no error, got %v", bn.Error())
		}
	})

	t.Run("with error", func(t *testing.T) {
		bn := NewBigN("invalid")
		if bn.Error() == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestToMoneyString(t *testing.T) {
	testCases := []struct {
		input       interface{}
		decimals    int32
		expected    string
		description string
	}{
		{NewBigN("1234567.89"), 2, "1,234,567.89", "Format 1234567.89 with 2 decimals"},
		{NewBigN("-9876543.21"), 2, "-9,876,543.21", "Format -9876543.21 with 2 decimals"},
		{NewBigN("1000"), 0, "1,000", "Format 1000 with 0 decimals"},
		{NewBigN("1234.5"), 2, "1,234.50", "Format 1234.5 with 2 decimals"},
		{NewBigN("0"), 2, "0.00", "Format 0 with 2 decimals"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.input.(*BigN).ToMoneyString(tc.decimals)
			if result != tc.expected {
				t.Errorf("ToMoneyString failed: got %v, want %v", result, tc.expected)
			}
		})
	}
}
