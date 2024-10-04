package transport

import (
	"net/http"

	"hw/pkg/micro-tree/endpoint"

	"github.com/gofiber/fiber/v2"
)

// DecodeRequestFunc defines how to decode data from an HTTP request.
type DecodeRequestFunc func(c *fiber.Ctx) (request interface{}, err error)

// EncodeResponseFunc defines how to encode the response into an HTTP response.
type EncodeResponseFunc func(c *fiber.Ctx, response interface{}) error

// Options contains the configuration options for the adapter.
type Options struct {
	Decode      DecodeRequestFunc
	Encode      EncodeResponseFunc
	Middlewares []fiber.Handler
}

// Option is a function type for setting Options.
type Option func(*Options)

// WithDecoder sets the decode function.
func WithDecoder(dec DecodeRequestFunc) Option {
	return func(opts *Options) {
		opts.Decode = dec
	}
}

// WithEncoder sets the encode function.
func WithEncoder(enc EncodeResponseFunc) Option {
	return func(opts *Options) {
		opts.Encode = enc
	}
}

// WithMiddleware adds an HTTP middleware.
func WithMiddleware(m fiber.Handler) Option {
	return func(opts *Options) {
		opts.Middlewares = append(opts.Middlewares, m)
	}
}

// New creates a new HTTP handler that adapts a go-kit style endpoint to an HTTP handler.
func New(ep endpoint.Endpoint, options ...Option) fiber.Handler {
	// Initialize options
	opts := &Options{}

	// Apply all provided options
	for _, option := range options {
		option(opts)
	}

	// Define the handler function
	handler := func(c *fiber.Ctx) error {
		// Check if endpoint is nil
		if ep == nil {
			return fiber.NewError(http.StatusInternalServerError, "endpoint cannot be nil")
		}

		// Get the request context
		ctx := c.Context()

		var request interface{}
		var err error

		// Use the decode function if provided
		if opts.Decode != nil {
			request, err = opts.Decode(c)
			if err != nil {
				return err
			}
		}

		// Call the endpoint with the request
		response, err := ep(ctx, request)
		if err != nil {
			return err
		}

		// Use the encode function if provided
		if opts.Encode != nil {
			return opts.Encode(c, response)
		}

		// If no encode function, send status OK with no content
		return c.SendStatus(http.StatusOK)
	}

	// Apply middlewares from last to first
	for i := len(opts.Middlewares) - 1; i >= 0; i-- {
		mw := opts.Middlewares[i]
		next := handler
		handler = func(c *fiber.Ctx) error {
			if err := mw(c); err != nil {
				return err
			}
			return next(c)
		}
	}

	return handler
}
