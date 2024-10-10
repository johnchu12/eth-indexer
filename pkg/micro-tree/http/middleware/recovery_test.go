package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newTestLogger creates a logger for testing purposes.
func newTestLogger(buffer *bytes.Buffer) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(buffer),
		zapcore.DebugLevel,
	)
	return zap.New(core)
}

// TestRecoveryMiddleware_PanicRecovery tests the recovery middleware's panic handling.
func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := newTestLogger(&logBuffer)

	recoveryMiddleware := RecoveryMiddleware(logger)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	r := chi.NewRouter()
	r.Use(recoveryMiddleware)
	r.Handle("/panic", panicHandler)

	req := httptest.NewRequest("GET", "/panic", nil)
	req = req.WithContext(context.WithValue(req.Context(), "requestid", "test-request-id"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	bodyBytes := new(bytes.Buffer)
	bodyBytes.ReadFrom(resp.Body)
	bodyString := bodyBytes.String()
	expectedBody := "Internal Server Error"
	if bodyString != expectedBody {
		t.Errorf("expected response body '%s', got '%s'", expectedBody, bodyString)
	}

	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "Panic occurred") {
		t.Errorf("expected log to contain 'Panic occurred', got '%s'", logOutput)
	}
	if !strings.Contains(logOutput, `"requestId":"test-request-id"`) {
		t.Errorf("expected log to contain requestId 'test-request-id', got '%s'", logOutput)
	}
	if !strings.Contains(logOutput, `"request":"/panic"`) {
		t.Errorf("expected log to contain request '/panic', got '%s'", logOutput)
	}
	if !strings.Contains(logOutput, `"error":"test panic"`) {
		t.Errorf("expected log to contain error 'test panic', got '%s'", logOutput)
	}
}

// TestRecoveryMiddleware_NoPanic tests the recovery middleware's behavior when no panic occurs.
func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := newTestLogger(&logBuffer)

	recoveryMiddleware := RecoveryMiddleware(logger)

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "OK")
	})

	r := chi.NewRouter()
	r.Use(recoveryMiddleware)
	r.Handle("/ok", okHandler)

	req := httptest.NewRequest("GET", "/ok", nil)
	req = req.WithContext(context.WithValue(req.Context(), "requestid", "test-request-id"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	bodyBytes := new(bytes.Buffer)
	bodyBytes.ReadFrom(resp.Body)
	bodyString := bodyBytes.String()
	expectedBody := "OK"
	if bodyString != expectedBody {
		t.Errorf("expected response body '%s', got '%s'", expectedBody, bodyString)
	}

	logOutput := logBuffer.String()
	if logOutput != "" {
		t.Errorf("expected no log output, got '%s'", logOutput)
	}
}

// TestLimitStackTrace tests the limitStackTrace function.
func TestLimitStackTrace(t *testing.T) {
	stack := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"
	limitedStack := limitStackTrace(stack, 3)
	lines := strings.Split(limitedStack, "\n")
	if len(lines) != 6 {
		t.Errorf("expected stack trace to have 6 lines, got %d lines", len(lines))
	}
	expectedLines := []string{"line1", "line2", "line3", "line4", "line5", "line6"}
	for i, line := range lines {
		if line != expectedLines[i] {
			t.Errorf("expected line %d to be '%s', got '%s'", i+1, expectedLines[i], line)
		}
	}
}

// TestGetRequestID tests the getRequestID function.
func TestGetRequestID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	reqID := getRequestID(req)
	if reqID != "unknown" {
		t.Errorf("expected request ID 'unknown', got '%s'", reqID)
	}

	ctx := context.WithValue(req.Context(), "requestid", "test-request-id")
	req = req.WithContext(ctx)
	reqID = getRequestID(req)
	if reqID != "test-request-id" {
		t.Errorf("expected request ID 'test-request-id', got '%s'", reqID)
	}
}
