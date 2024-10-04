package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) fiber.Handler {
	cfg := ConfigDefault
	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		if err := c.Next(); err != nil {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// record error message
			logger.Error("error handler:",
				zap.String("requestId", getRequestID(c)),
				zap.Int("status", code),
				zap.String("path", c.Path()),
				zap.String("ip", getClientIP(c)),
				zap.Error(err),
			)

			switch code {
			case http.StatusBadRequest:
				return c.Status(code).SendString("400 Bad Request")
			case http.StatusUnauthorized:
				return c.Status(code).SendString("401 Unauthorized")
			case http.StatusForbidden:
				return c.Status(code).SendString("403 Forbidden")
			case http.StatusNotFound:
				return c.Status(code).SendString("404 Not Found")
			case http.StatusInternalServerError:
				return c.Status(code).SendString("500 Internal Server Error")
			default:
				return c.Status(code).SendString(err.Error())
			}
		}

		return nil
	}
}

func getRequestID(c *fiber.Ctx) string {
	id := c.Locals("requestid")
	if id != nil {
		return id.(string)
	}
	return "unknown"
}
