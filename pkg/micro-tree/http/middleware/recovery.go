package middleware

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RecoveryConfig defines the config for recovery middleware
type RecoveryConfig struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c *fiber.Ctx) bool

	// Logger is the zap logger instance
	Logger *zap.Logger
}

// RecoveryConfigDefault is the default recovery middleware config
var RecoveryConfigDefault = RecoveryConfig{
	Next:   nil,
	Logger: nil,
}

// Recovery returns a Fiber middleware which recovers from panics anywhere in the stack
// and handles the control to the centralized ErrorHandler.
func Recovery(log *zap.Logger) fiber.Handler {
	// Set default config
	cfg := RecoveryConfigDefault

	cfg.Logger = log

	// Return new handler
	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				stack := string(debug.Stack())
				clientIP := getClientIP(c)

				log.Error("Panic occurred",
					zap.String("requestId", c.Locals("requestid").(string)),
					zap.String("request", c.OriginalURL()),
					zap.String("ip", clientIP),
					zap.Any("error", err),
					zap.String("stacktrace", limitStackTrace(stack, 6)),
				)

				// errHandler := c.App().Config().ErrorHandler
				// if errHandler != nil {
				// 	errHandler(c, err)
				// } else {
				// 	// Default error handling
				// 	c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
				// }
				c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
		}()
		return c.Next()
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
