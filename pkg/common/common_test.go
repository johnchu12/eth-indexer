package common_test

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
	"time"

	"hw/pkg/common"

	"github.com/stretchr/testify/assert"
)

// TestGetEnv tests the GetEnv function
func TestGetEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	// Test the set environment variable
	value := common.GetEnv("TEST_KEY", "default")
	assert.Equal(t, "test_value", value, "should return the set environment variable value")

	// Test an unset environment variable
	value = common.GetEnv("NON_EXISTENT_KEY", "default")
	assert.Equal(t, "default", value, "should return the default value")
}

// TestMakeCurrentTimestamp tests the MakeCurrentTimestamp function
func TestMakeCurrentTimestamp(t *testing.T) {
	timestamp := common.MakeCurrenctTimestamp()
	current := time.Now().UnixNano() / int64(time.Millisecond)
	assert.InDelta(t, current, timestamp, 10, "the timestamp should be close to the current time")
}

// TestMustParseDuration tests the MustParseDuration function
func TestMustParseDuration(t *testing.T) {
	duration := common.MustParseDuration("2h45m")
	expected := 2*time.Hour + 45*time.Minute
	assert.Equal(t, expected, duration, "should correctly parse the duration")

	// Test an invalid duration, should panic
	assert.Panics(t, func() { common.MustParseDuration("invalid") }, "should panic when parsing an invalid duration")
}

// TestPrintStruct tests the PrintStruct function
func TestPrintStruct(t *testing.T) {
	input := map[string]interface{}{
		"key": "value",
	}
	expected := `{"key":"value"}`
	result := common.PrintStruct(input)
	assert.JSONEq(t, expected, result, "should correctly serialize the struct to a JSON string")
}

// TestPrintStructRaw tests the PrintStructRaw function
func TestPrintStructRaw(t *testing.T) {
	input := map[string]interface{}{
		"key": "value",
	}
	expected := json.RawMessage(`{"key":"value"}`)
	result := common.PrintStructRaw(input)
	assert.Equal(t, expected, result, "should correctly serialize the struct to json.RawMessage")
}

// TestGetRegexMap tests the GetRegexMap function
func TestGetRegexMap(t *testing.T) {
	re := regexp.MustCompile(`(?P<first>\w+) (?P<second>\w+)`)
	msg := "hello world"
	expected := map[string]string{
		"first":  "hello",
		"second": "world",
	}
	result := common.GetRegexMap(re, msg)
	assert.Equal(t, expected, result, "should correctly extract the regex named group matches")
}

// TestFormatHashrate tests the FormatHashrate function
func TestFormatHashrate(t *testing.T) {
	tests := []struct {
		hashrate string
		decimal  int32
		expected string
	}{
		{"1234567890123", 2, "1.23T"},
		{"1234567890", 2, "1.23G"},
		{"1234567", 2, "1.23M"},
		{"1234", 2, "1.23K"},
		{"123", 2, "123.00"},
	}

	for _, tt := range tests {
		result := common.FormatHashrate(tt.hashrate, tt.decimal)
		assert.Equal(t, tt.expected, result, "should correctly format the hashrate")
	}
}
