package request

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// TestClient_Do tests the Do method of the Client.
func TestClient_Do(t *testing.T) {
	// Initialize a test server with different endpoints to simulate various scenarios.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		case "/post":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "created"}`))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		case "/timeout":
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize the client with default settings and base URL pointing to the test server.
	client := NewClient(
		BaseURL(server.URL),
		Timeout("3s"),
		Logger(logger.Sugar()),
		SetRetryCount(1),
	)

	// Test case: Successful GET request
	t.Run("Successful GET Request", func(t *testing.T) {
		resp, err := client.Do("GET", "/success")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
		expectedBody := `{"message": "success"}`
		if string(resp.Data) != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
		}
	})

	// Test case: Successful POST request
	t.Run("Successful POST Request", func(t *testing.T) {
		resp, err := client.Do("POST", "/post")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}
		expectedBody := `{"message": "created"}`
		if string(resp.Data) != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
		}
	})

	// Test case: Unsupported HTTP method
	t.Run("Unsupported HTTP Method", func(t *testing.T) {
		_, err := client.Do("PUT", "/success")
		if err == nil {
			t.Fatalf("Expected error for unsupported method, got nil")
		}
		expectedErr := "unsupported method: PUT"
		if err.Error() != expectedErr {
			t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
		}
	})

	// Test case: Server returns an error response
	t.Run("Server Error Response", func(t *testing.T) {
		resp, err := client.Do("GET", "/error")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
		expectedBody := `{"error": "internal server error"}`
		if string(resp.Data) != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
		}
	})
}

// isTimeoutError checks if the error is a timeout error.
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	switch e := err.(type) {
	case net.Error:
		return e.Timeout()
	case interface{ Timeout() bool }:
		return e.Timeout()
	default:
		return false
	}
}

// TestClient_Do_WithHeaders tests setting custom headers.
func TestClient_Do_WithHeaders(t *testing.T) {
	// Initialize test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "CustomValue" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "header received"}`))
	}))
	defer server.Close()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize client with custom headers
	client := NewClient(
		BaseURL(server.URL),
		Header(map[string]string{
			"X-Custom-Header": "CustomValue",
		}),
		Logger(logger.Sugar()),
	)

	// Execute request
	resp, err := client.Do("GET", "/")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	expectedBody := `{"message": "header received"}`
	if string(resp.Data) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
	}
}

// TestClient_Do_WithAuthToken tests setting an authentication token.
func TestClient_Do_WithAuthToken(t *testing.T) {
	// Initialize test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer testtoken" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "authorized"}`))
	}))
	defer server.Close()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize client with authentication token
	client := NewClient(
		BaseURL(server.URL),
		AuthToken("testtoken"),
		Logger(logger.Sugar()),
	)

	// Execute request
	resp, err := client.Do("GET", "/")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	expectedBody := `{"message": "authorized"}`
	if string(resp.Data) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
	}
}

// TestClient_Do_WithSetFormData tests setting form data.
func TestClient_Do_WithSetFormData(t *testing.T) {
	// Initialize test server to verify form data is received correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.FormValue("field1") != "value1" || r.FormValue("field2") != "value2" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "form data received"}`))
	}))
	defer server.Close()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize client and set form data
	client := NewClient(
		BaseURL(server.URL),
		Logger(logger.Sugar()),
	)
	client.SetFormData(map[string]string{
		"field1": "value1",
		"field2": "value2",
	})

	// Execute request
	resp, err := client.Do("POST", "/")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	expectedBody := `{"message": "form data received"}`
	if string(resp.Data) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
	}
}

// TestClient_Do_WithSetBody tests setting the request body.
func TestClient_Do_WithSetBody(t *testing.T) {
	type RequestBody struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Initialize test server to verify request body is received correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body RequestBody
		if err := jsoniter.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Name != "John Doe" || body.Email != "john@example.com" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "body received"}`))
	}))
	defer server.Close()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize client and set request body
	client := NewClient(
		BaseURL(server.URL),
		Logger(logger.Sugar()),
	)
	client.SetBody(RequestBody{
		Name:  "John Doe",
		Email: "john@example.com",
	})

	// Execute request
	resp, err := client.Do("POST", "/")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	expectedBody := `{"message": "body received"}`
	if string(resp.Data) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
	}
}

// TestClient_Do_WithUnsupportedMethod tests behavior when using an unsupported HTTP method.
func TestClient_Do_WithUnsupportedMethod(t *testing.T) {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize client
	client := NewClient(
		BaseURL("http://example.com"),
		Logger(logger.Sugar()),
	)

	// Execute request with unsupported method
	_, err = client.Do("DELETE", "/")
	if err == nil {
		t.Fatalf("Expected error for unsupported method, got nil")
	}
	expectedErr := "unsupported method: DELETE"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestClient_Do_WithErrorHandler tests setting an error handler.
func TestClient_Do_WithErrorHandler(t *testing.T) {
	// Define custom error structure
	type ErrorResponse struct {
		Error string `json:"error"`
	}

	// Initialize test server to return an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize client and set error handler
	client := NewClient(
		BaseURL(server.URL),
		SetErrorHandler(&ErrorResponse{}),
		Logger(logger.Sugar()),
	)

	// Execute request
	resp, err := client.Do("GET", "/")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	expectedBody := `{"error": "bad request"}`
	if string(resp.Data) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Data))
	}
}

// TestNewClient_DefaultSettings tests the default settings of NewClient.
func TestNewClient_DefaultSettings(t *testing.T) {
	client := NewClient()

	// Check if the default timeout is 18 seconds
	if client.client.GetClient().Timeout != 18*time.Second {
		t.Errorf("Expected default timeout to be 18s, got %v", client.client.GetClient().Timeout)
	}

	// Check if the default retry count is 3
	if client.client.RetryCount != 3 {
		t.Errorf("Expected default RetryCount to be 3, got %d", client.client.RetryCount)
	}

	// Check if the default retry wait time is 2 seconds
	if client.client.RetryWaitTime != 2*time.Second {
		t.Errorf("Expected default RetryWaitTime to be 2s, got %v", client.client.RetryWaitTime)
	}

	// Check if the default maximum retry wait time is 1 minute
	if client.client.RetryMaxWaitTime != time.Minute {
		t.Errorf("Expected default RetryMaxWaitTime to be 1m, got %v", client.client.RetryMaxWaitTime)
	}
}
