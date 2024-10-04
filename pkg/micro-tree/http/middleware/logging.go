package middleware

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Config defines the config for logger middleware
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c *fiber.Ctx) bool

	// Logger is the zap logger instance
	Logger *zap.Logger

	// TimeFormat defines the time format for log timestamps.
	TimeFormat string

	// TimeZone can be specified, such as "UTC" and "America/New_York" and "Asia/Shanghai", etc
	TimeZone string

	// Output is the io.Writer to write the log message to
	Output io.Writer
}

// ConfigDefault is the default configuration
var ConfigDefault = Config{
	Next:       nil,
	Logger:     nil,
	TimeFormat: "2006-01-02 15:04:05",
	TimeZone:   "Local",
	Output:     os.Stdout,
}

// UseLog is a helper function to quickly set up the logger middleware
func UseLog(app *fiber.App, log *zap.Logger) {
	app.Use(New(Config{
		Logger:     log,
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.Next == nil {
			cfg.Next = ConfigDefault.Next
		}

		if cfg.TimeFormat == "" {
			cfg.TimeFormat = ConfigDefault.TimeFormat
		}

		if cfg.TimeZone == "" {
			cfg.TimeZone = ConfigDefault.TimeZone
		}

		if cfg.Output == nil {
			cfg.Output = ConfigDefault.Output
		}
	}

	// Get timezone location
	tz, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		panic(err)
	}

	// Return new handler
	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Handle RequestID
		spanID := getSpanIDFromContext(c)
		var requestID string
		if spanID != "00000000000000000000000000000000" {
			requestID = spanID
		} else {
			requestID = "11" + uuid.New().String()[:8]
		}

		c.Locals("requestid", requestID)

		// Set variables
		start := time.Now().In(tz)
		path := c.Path()
		method := c.Method()

		// Handle request
		chainErr := c.Next()

		// Set latency
		stop := time.Now().In(tz)
		latency := stop.Sub(start)

		// Get status code
		status := c.Response().StatusCode()

		// Get client IP
		clientIP := getClientIP(c)

		// Get user agent
		userAgent := c.Get(fiber.HeaderUserAgent)
		if userAgent == "" {
			userAgent = "unknown agent"
		}

		// Format log message
		logMessage := fmt.Sprintf("%s %d %s %s %s %s %.3fKB %s \"%s\"",
			method,
			status,
			timeDurationFormat(latency),
			requestID,
			clientIP,
			path,
			float64(c.Response().Header.ContentLength())/1024.0,
			c.Protocol(),
			userAgent,
		)

		if chainErr != nil {
			logMessage = fmt.Sprintf("%s %s", logMessage, chainErr.Error())
		}

		// Log message
		if cfg.Logger != nil {
			cfg.Logger.Info(logMessage)
		} else {
			_, _ = cfg.Output.Write([]byte(logMessage + "\n"))
		}

		return chainErr
	}
}

func GetClientIP(c *fiber.Ctx) string {
	return getClientIP(c)
}

func getClientIP(c *fiber.Ctx) string {
	// First try to get IP from CF-Connecting-IP header
	clientIP := c.Get("CF-Connecting-IP")
	if clientIP != "" {
		return clientIP
	}
	// If not found, try X-Forwarded-For header
	clientIP = c.Get("X-Forwarded-For")
	if clientIP != "" {
		return strings.Split(clientIP, ",")[0] // X-Forwarded-For may contain multiple IPs, we only need the first one
	}
	// If all else fails, return the RemoteIP
	return c.IP()
}

func timeDurationFormat(t time.Duration) string {
	totalSeconds := t.Seconds()

	if totalSeconds >= 1 {
		// Greater than or equal to 1 second, show seconds with three decimal places
		return fmt.Sprintf("%.3fs", totalSeconds)
	} else if totalSeconds >= 0.001 {
		// Greater than or equal to 1 millisecond, show milliseconds with three decimal places
		return fmt.Sprintf("%.3fms", totalSeconds*1000)
	}

	// For other cases, use the default string representation
	return t.String()
}

// getSpanIDFromContext extracts the span ID from the request context
func getSpanIDFromContext(c *fiber.Ctx) string {
	// Get the context from Fiber's fasthttp.RequestCtx
	ctx := c.Context()

	// Extract the span from the context
	span := trace.SpanFromContext(ctx)

	// Return the trace ID as a string
	return span.SpanContext().TraceID().String()
}
