package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Unic-X/slow-server/api"
	"github.com/Unic-X/slow-server/config"
	"github.com/Unic-X/slow-server/models"
)

func setupTestConfig() *config.Config {
	// Create test config with predictable behavior
	testCfg := &config.Config{
		Port:           8080,
		LogLevel:       "info",
		SimulateErrors: false, // Disable errors for deterministic tests
		MinDelay:       10,    // Short delays for faster tests
		MaxDelay:       50,
		DBQueryDelay:   20,
		APICallDelay:   30,
		ProcessDelay:   15,
		ErrorRate:      0,
		EnableMetrics:  true,
	}
	// Set the test config in the API package
	api.SetConfig(testCfg)
	return testCfg
}

func TestGetDataEndpoint(t *testing.T) {
	setupTestConfig()

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/api/data", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add a request ID header
	req.Header.Set("X-Request-ID", "test-request-id")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.GetDataHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body has the expected format
	var response models.DataResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Verify the response contains data
	if len(response.Data) == 0 {
		t.Errorf("Expected data items in response, got empty array")
	}

	// Verify the count matches the actual data length
	if response.Count != len(response.Data) {
		t.Errorf("Response count doesn't match data length: got %v want %v",
			response.Count, len(response.Data))
	}
}

func TestGetUsersEndpoint(t *testing.T) {
	setupTestConfig()

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add a request ID header
	req.Header.Set("X-Request-ID", "test-request-id")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.GetUsersHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body has the expected format
	var response models.UsersResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Verify the response contains users
	if len(response.Users) == 0 {
		t.Errorf("Expected users in response, got empty array")
	}

	// Verify the count matches the actual users length
	if response.Count != len(response.Users) {
		t.Errorf("Response count doesn't match users length: got %v want %v",
			response.Count, len(response.Users))
	}
}

func TestProcessDataEndpoint(t *testing.T) {
	setupTestConfig()

	// Create request body
	requestBody := `{
		"items": ["item1", "item2", "item3"],
		"args": {
			"param1": "value1",
			"param2": "value2"
		}
	}`

	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/api/process", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Set content type and request ID
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.ProcessDataHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body has the expected format
	var response models.ProcessResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Verify the response indicates success
	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}

	// Verify the process ID is set
	if response.ProcessID == 0 {
		t.Errorf("Expected non-zero process ID, got 0")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	setupTestConfig()

	// Test POST on a GET endpoint
	req, err := http.NewRequest("POST", "/api/data", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.GetDataHandler)
	handler.ServeHTTP(rr, req)

	// Check that method is not allowed
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code for wrong method: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestInvalidRequestBody(t *testing.T) {
	setupTestConfig()

	// Create invalid JSON request body
	requestBody := `{
		"items": ["item1", "item2"
		"args": {
			"param1": "value1"
		}
	}` // Missing comma, invalid JSON

	req, err := http.NewRequest("POST", "/api/process", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.ProcessDataHandler)
	handler.ServeHTTP(rr, req)

	// Check that the request is considered bad
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler should reject invalid JSON with 400 Bad Request: got %v", status)
	}
}

func TestResponseTime(t *testing.T) {
	setupTestConfig()

	req, err := http.NewRequest("GET", "/api/data", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.GetDataHandler)

	// Measure response time
	start := time.Now()
	handler.ServeHTTP(rr, req)
	duration := time.Since(start)

	// The test config has delays, so we expect some minimum response time
	// MinDelay + DBQueryDelay + ProcessDelay would be the minimum expected
	// But it could be more due to actual processing time
	minExpectedDuration := 45 * time.Millisecond // 10 + 20 + 15

	if duration < minExpectedDuration {
		t.Errorf("Response was too fast: %v, expected at least %v", 
			duration, minExpectedDuration)
	}

	// Also check we got a valid response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
