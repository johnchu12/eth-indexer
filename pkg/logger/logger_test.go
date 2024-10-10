package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogEntry represents the structure of a single log entry as a map.
type LogEntry map[string]interface{}

// setupTestLogger initializes a test logger that writes to a buffer using a JSON encoder for parsing.
func setupTestLogger() (*zap.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zap.DebugLevel,
	)
	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
	return logger, &buf
}

// parseLogEntry parses a single log line into a LogEntry structure.
func parseLogEntry(t *testing.T, logLine string) LogEntry {
	var entry LogEntry
	err := json.Unmarshal([]byte(logLine), &entry)
	assert.NoError(t, err, "Failed to unmarshal log entry")
	return entry
}

func TestInfow(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	Infow("Test Infow", "key1", "value1")

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	// Split each log line into separate entries
	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "Test Infow", entry["msg"])
	value, exists := entry["key1"]
	assert.True(t, exists, "key1 should exist")
	assert.Equal(t, "value1", value)
}

func TestInfof(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	Infof("Test Infof with %s", "formatted string")

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "Test Infof with formatted string", entry["msg"])
}

func TestWarnw(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	Warnw("Test Warnw", "warnKey", "warnValue")

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "Test Warnw", entry["msg"])
	value, exists := entry["warnKey"]
	assert.True(t, exists, "warnKey should exist")
	assert.Equal(t, "warnValue", value)
}

func TestWarnf(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	Warnf("Test Warnf with number %d", 42)

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "Test Warnf with number 42", entry["msg"])
}

func TestErrorw(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	Errorw("Test Errorw", "errorKey", "errorValue")

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "Test Errorw", entry["msg"])
	value, exists := entry["errorKey"]
	assert.True(t, exists, "errorKey should exist")
	assert.Equal(t, "errorValue", value)
	stacktrace, exists := entry["stacktrace"]
	assert.True(t, exists, "stacktrace should exist")
	assert.NotEmpty(t, stacktrace, "Stacktrace should be present")
}

func TestErrorf(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	Errorf("Test Errorf with value %v", 3.14)

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "Test Errorf with value 3.14", entry["msg"])
}

func TestErrorfWithStack(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	ErrorfWithStack("Test ErrorfWithStack")

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "Test ErrorfWithStack", entry["msg"])
	stacktrace, exists := entry["stacktrace"]
	assert.True(t, exists, "stacktrace should exist")
	assert.NotEmpty(t, stacktrace, "Stacktrace should be present")
}

func TestInfoField(t *testing.T) {
	logger, buf := setupTestLogger()
	defer logger.Sync()

	// Act
	fields := []zapcore.Field{
		zap.String("field1", "value1"),
		zap.Int("field2", 2),
	}
	InfoField("Test InfoField", fields...)

	// Assert
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput, "Log output should not be empty")

	logLines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, logLines, 2, "Should have one log entry")

	entry := parseLogEntry(t, string(logLines[0]))
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "Test InfoField", entry["msg"])

	// Verify field1
	field1, exists := entry["field1"]
	assert.True(t, exists, "field1 should exist")
	assert.Equal(t, "value1", field1)

	// Verify field2
	field2, exists := entry["field2"]
	assert.True(t, exists, "field2 should exist")
	assert.Equal(t, float64(2), field2) // JSON numbers are float64
}

func TestInit(t *testing.T) {
	logger := Init()
	defer logger.Sync()

	// Assert
	assert.NotNil(t, logger, "Logger should not be nil")
	assert.Equal(t, zap.L(), logger, "Global logger should be equal to the initialized logger")
}
