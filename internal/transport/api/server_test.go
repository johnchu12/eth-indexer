package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hw/internal/service/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// setupTestRouter initializes the test router.
func setupTestRouter(srv Server) *chi.Mux {
	router := chi.NewRouter()
	ConfigureHTTPServer(router, srv)
	return router
}

// TestPingHandler tests the /ping route.
func TestPingHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := mocks.NewMockService(gomock.NewController(t))
	srv := Server{
		Logger:  logger,
		Service: mockService,
	}
	router := setupTestRouter(srv)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := "pong"
	if w.Body.String() != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
	}
}

// TestErrorHandler tests the /error route.
func TestErrorHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := mocks.NewMockService(gomock.NewController(t))
	srv := Server{
		Logger:  logger,
		Service: mockService,
	}
	router := setupTestRouter(srv)

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var errResp errorResponse
	if err := render.DecodeJSON(w.Body, &errResp); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	expectedError := "errtest"
	if errResp.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errResp.Error)
	}
}

// TestPanicHandler tests the /panic route.
func TestPanicHandler(t *testing.T) {
	// Initialize logger with development config to capture panic logs
	config := zap.NewDevelopmentConfig()

	config.OutputPaths = []string{"stdout"}
	logger, _ := config.Build()
	defer logger.Sync()

	mockService := mocks.NewMockService(gomock.NewController(t))
	srv := Server{
		Logger:  logger,
		Service: mockService,
	}
	router := setupTestRouter(srv)

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Since panic is recovered, expect 500 Internal Server Error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
