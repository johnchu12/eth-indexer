package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RequestIDMiddleware returns a Chi middleware for generating and setting a Request ID.
func RequestIDMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate and set the Request ID
			requestID := generateRequestID(r)
			ctx := r.Context()
			ctx = contextWithRequestID(ctx, requestID)
			r = r.WithContext(ctx)

			// Optionally set the Request ID in the response header
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware returns a Chi middleware for logging requests upon completion.
func LoggingMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate and set the Request ID
			requestID := generateRequestID(r)
			ctx := r.Context()
			ctx = contextWithRequestID(ctx, requestID)
			r = r.WithContext(ctx)

			// Wrap the ResponseWriter to capture status code and response size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			// Execute the next handler
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			// Extract necessary request information
			method := r.Method
			status := ww.Status()
			path := r.URL.Path
			proto := r.Proto
			userAgent := r.UserAgent()
			if userAgent == "" {
				userAgent = "unknown agent"
			}
			clientIP := getClientIP(r)
			// Note: ww.BytesWritten() returns bytes, not KB
			responseSizeKB := float64(ww.BytesWritten()) / 1024.0

			// Extract error information (ServeHTTP does not return errors, so keep it empty)
			errText := ""

			// Log the request
			logger.Info(fmt.Sprintf("%s %d %s %s %s %s %.3fKB %s \"%s\" %v",
				method,
				status,
				timeDurationFormat(duration),
				requestID,
				clientIP,
				path,
				responseSizeKB, // Note: this is not the response body size
				proto,
				userAgent,
				errText,
			))
		})
	}
}

// generateRequestID generates a Request ID, preferring Span ID if available, otherwise generating a UUID.
func generateRequestID(r *http.Request) string {
	spanID := getSpanIDFromContext(r)
	if spanID != "" {
		return spanID
	}
	return "11" + uuid.New().String()[:8]
}

// contextWithRequestID sets the request ID in the context.
func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "requestid", requestID)
}

// getSpanIDFromContext extracts the Span ID from the request context.
func getSpanIDFromContext(r *http.Request) string {
	span := trace.SpanFromContext(r.Context())
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// getClientIP extracts the client's IP address.
func getClientIP(r *http.Request) string {
	// First attempt to get IP from CF-Connecting-IP header
	clientIP := r.Header.Get("CF-Connecting-IP")
	if clientIP != "" {
		return clientIP
	}
	// If the above header is not present, attempt to get from X-Forwarded-For
	clientIP = r.Header.Get("X-Forwarded-For")
	if clientIP != "" {
		// X-Forwarded-For may contain multiple IPs, take the first one
		return strings.Split(clientIP, ",")[0]
	}
	// If both methods fail, return RemoteAddr
	ip := r.RemoteAddr
	// Remove port number
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		return ip[:colon]
	}
	return ip
}

// timeDurationFormat formats the duration to display in seconds or milliseconds.
func timeDurationFormat(t time.Duration) string {
	totalSeconds := t.Seconds()

	if totalSeconds >= 1 {
		// Display seconds with three decimal places
		return fmt.Sprintf("%.3fs", totalSeconds)
	} else if totalSeconds >= 0.001 {
		// Display milliseconds with three decimal places
		return fmt.Sprintf("%.3fms", totalSeconds*1000)
	}

	// Otherwise, use the default string representation
	return t.String()
}
