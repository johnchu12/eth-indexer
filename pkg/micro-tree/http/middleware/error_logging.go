package middleware

import (
	"hw/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HTTPLogging is a middleware function that logs HTTP request details
func HTTPErrorLogging(c *fiber.Ctx, err error) {
	// Retrieve the request ID from the context
	requestIDFromLocals := c.Locals("requestid")

	// Initialize log fields with common request information
	logFields := []zap.Field{
		zap.String("req_id", requestIDFromLocals.(string)),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Any("query_params", c.Queries()),
	}

	// For POST requests, log form data and/or request body
	if c.Method() == fiber.MethodPost {
		// Attempt to parse and log multipart form data
		form, err := c.MultipartForm()
		if err == nil {
			logFields = append(logFields, zap.Any("form_data", form.Value))
		}

		// Retrieve the request body
		body := c.Body()

		// If the body is not empty, log its contents
		if len(body) > 0 {
			logFields = append(logFields, zap.String("body", string(body)))
		}
	}

	// If there's an error, add it to the log fields
	if err != nil {
		logFields = append(logFields, zap.String("error", err.Error()))
	}

	// Log the request details using the configured logger
	logger.InfoField("request", logFields...)
}
