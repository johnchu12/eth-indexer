package endpoint

import (
	"context"
)

// Endpoint defines a generic service function type.
// It accepts a context and a request interface, returning a response interface and an error.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Nop is a no-operation Endpoint implementation that always returns an empty struct and a nil error.
// It can be used for testing or as a placeholder.
func Nop(ctx context.Context, request interface{}) (interface{}, error) {
	return struct{}{}, nil
}

// Middleware defines a middleware function type.
// It takes an Endpoint and returns a new Endpoint, allowing additional logic before and after the original service.
type Middleware func(Endpoint) Endpoint

// Chain is used to combine multiple middlewares into one.
// It accepts an outer middleware and any number of additional middlewares, returning a new combined middleware.
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(others) - 1; i >= 0; i-- { // Iterate in reverse
			next = others[i](next)
		}
		return outer(next)
	}
}

// Failer interface defines a Failed method to check if an operation has failed.
// Types implementing this interface can provide more detailed error information.
type Failer interface {
	Failed() error
}
