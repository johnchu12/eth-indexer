package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// RecoveryConfig defines the config for recovery middleware
type RecoveryConfig struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(w http.ResponseWriter, r *http.Request) bool

	// Logger is the zap logger instance
	Logger *zap.Logger
}

// RecoveryConfigDefault is the default recovery middleware config
var RecoveryConfigDefault = RecoveryConfig{
	Next:   nil,
	Logger: nil,
}

// RecoveryMiddleware returns a Fiber middleware which recovers from panics anywhere in the stack
// and handles the control to the centralized ErrorHandler.
func RecoveryMiddleware(log *zap.Logger) func(http.Handler) http.Handler {
	// Set default config
	cfg := RecoveryConfigDefault

	cfg.Logger = log

	// Return new handler
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't execute middleware if Next returns true
			if cfg.Next != nil && cfg.Next(w, r) {
				next.ServeHTTP(w, r)
				return
			}

			defer func() {
				if rec := recover(); rec != nil {
					var err error
					if e, ok := rec.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("%v", rec)
					}
					stack := string(debug.Stack())
					clientIP := getClientIP(r)

					cfg.Logger.Error("Panic occurred",
						zap.String("requestId", getRequestID(r)),
						zap.String("request", r.URL.Path),
						zap.String("ip", clientIP),
						zap.Any("error", err),
						zap.String("stacktrace", limitStackTrace(stack, 6)),
					)

					w.WriteHeader(http.StatusInternalServerError)
					render.PlainText(w, r, "Internal Server Error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// limitStackTrace limits the number of stack trace levels
func limitStackTrace(stack string, limit int) string {
	lines := strings.Split(stack, "\n")
	displayLines := limit * 2 // Each call usually takes two lines
	if displayLines > len(lines) {
		displayLines = len(lines)
	}
	return strings.Join(lines[:displayLines], "\n")
}

// getRequestID gets the request's ID
func getRequestID(r *http.Request) string {
	requestID := r.Context().Value("requestid")
	if requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}
