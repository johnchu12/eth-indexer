package middleware

import (
	"bytes"
	"io"
	"net/http"

	"hw/pkg/logger"

	"go.uber.org/zap"
)

// HTTPErrorLogging is a middleware function that logs HTTP request details including errors.
func HTTPErrorLogging(w http.ResponseWriter, r *http.Request, err error) {
	// Retrieve the request ID from the context
	requestID := r.Context().Value("requestid").(string)

	// Initialize log fields with common request information
	logFields := []zap.Field{
		zap.String("req_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("query_params", r.URL.Query()),
	}

	// For POST requests, log form data and/or request body
	if r.Method == http.MethodPost {
		// Attempt to parse and log multipart form data
		if parseErr := r.ParseForm(); parseErr == nil {
			logFields = append(logFields, zap.Any("form", r.Form))
		}

		// Retrieve the request body
		bodyBytes, readErr := io.ReadAll(r.Body)
		if readErr == nil && len(bodyBytes) > 0 {
			// Restore the read body data back to the request for downstream handlers
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Log body data
			logFields = append(logFields, zap.String("body", string(bodyBytes)))
		}
	}

	// If there's an error, add it to the log fields
	if err != nil {
		logFields = append(logFields, zap.String("error", err.Error()))
	}

	// Log the request details using the configured logger
	logger.GetLogger().With(logFields...).Error("request")
}
