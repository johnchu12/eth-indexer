package server

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// TestNewHTTPServer tests the NewHTTPServer function.
func TestNewHTTPServer(t *testing.T) {
	router := NewHTTPServer()
	assert.NotNil(t, router, "Router should not be nil")
	assert.IsType(t, &chi.Mux{}, router, "Should return an instance of *chi.Mux type")
}
