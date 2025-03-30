package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Unic-X/slow-server/models"
)

// Integration tests that run against a live server instance

const (
	serverURL = "http://localhost:8080"
	timeout   = 30 * time.Second
)

var serverProcess *exec.Cmd

// TestMain sets up the test environment by starting the server
func TestMain(m *testing.M) {
	// Only run the server startup in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		// Run the tests without starting the server
		os.Exit(m.Run())
		return
	}

	// Set environment variables for deterministic testing
	os.Setenv("SIMULATE_ERRORS", "false")
	os.Setenv("MIN_DELAY", "100")
	os.Setenv("MAX_DELAY", "300")
	os.Setenv("SERVER_PORT", "8080")

	// Build the server
	buildCmd := exec.Command("go", "build", "-o", "test-server", "../")
	if err := buildCmd.Run(); err != nil {
		panic("Failed to build server: " + err.Error())
	}

	// Start the server
	serverProcess = exec.Command("./test-server")
	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr
	if err := serverProcess.Start(); err != nil {
		panic("Failed to start server: " + err.Error())
	}

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Run the tests
	result := m.Run()

	// Cleanup
	if serverProcess != nil && serverProcess.Process != nil {
		serverProcess.Process.Kill()
	}
	os.Remove("test-server")

	os.Exit(result)
}

// waitForServer waits for the server to become available
func waitForServer(t *testing.T) {
	// Skip this in unit test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test in unit test mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Server did not become available within %s", timeout)
		case <-ticker.C:
			resp, err := http.Get(serverURL + "/api/data")
			if err == nil {
				resp.Body.Close()
				return
			}
		}
	}
}

// Integration test for the data endpoint
func TestIntegrationGetData(t *testing.T) {
	// Ensure server is running
	waitForServer(t)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make the request
	resp, err := client.Get(serverURL + "/api/data")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var dataResp models.DataResponse
	err = json.NewDecoder(resp.Body).Decode(&dataResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check data
	if len(dataResp.Data) == 0 {
		t.Error("Expected data in response, got empty array")
	}

	// Check count
	if dataResp.Count != len(dataResp.Data) {
		t.Errorf("Expected count %d, got %d", len(dataResp.Data), dataResp.Count)
	}
}

// Integration test for the users endpoint
func TestIntegrationGetUsers(t *testing.T) {
	// Ensure server is running
	waitForServer(t)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make the request
	resp, err := client.Get(serverURL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var usersResp models.UsersResponse
	err = json.NewDecoder(resp.Body).Decode(&usersResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check users
	if len(usersResp.Users) == 0 {
		t.Error("Expected users in response, got empty array")
	}

	// Check count
	if usersResp.Count != len(usersResp.Users) {
		t.Errorf("Expected count %d, got %d", len(usersResp.Users), usersResp.Count)
	}
}

// Integration test for the process endpoint
func TestIntegrationProcessData(t *testing.T) {
	// Ensure server is running
	waitForServer(t)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second, // Longer timeout for the process endpoint
	}

	// Create request body
	requestBody := models.ProcessRequest{
		Items: []string{"item1", "item2", "item3"},
		Args: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Make the request
	resp, err := client.Post(
		serverURL+"/api/process",
		"application/json",
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var processResp models.ProcessResponse
	err = json.NewDecoder(resp.Body).Decode(&processResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check success
	if !processResp.Success {
		t.Error("Expected success to be true, got false")
	}

	// Check process ID
	if processResp.ProcessID == 0 {
		t.Error("Expected non-zero process ID, got 0")
	}
}

// Test response time for multiple endpoints
func TestIntegrationResponseTimes(t *testing.T) {
	// Ensure server is running
	waitForServer(t)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	endpoints := []string{"/api/data", "/api/users"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			start := time.Now()
			resp, err := client.Get(serverURL + endpoint)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", endpoint, err)
			}
			defer resp.Body.Close()
			
			duration := time.Since(start)
			
			// Check that response time is within expected range
			if duration < 100*time.Millisecond {
				t.Errorf("Response was too fast: %v, expected at least 100ms", duration)
			}
			
			// Also check the response is valid
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
			}
			
			// Log the response time for information
			t.Logf("Response time for %s: %v", endpoint, duration)
		})
	}
}

// Test error simulation by enabling errors during the test
func TestIntegrationWithErrors(t *testing.T) {
	// Skip this specific test if not in integration mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping error simulation test in unit test mode")
		return
	}

	// Save the current error rate
	prevErrorRate := os.Getenv("ERROR_RATE")
	
	// Set a high error rate
	os.Setenv("ERROR_RATE", "0.8") // 80% error rate
	os.Setenv("SIMULATE_ERRORS", "true")
	
	// Restart the server with new settings
	if serverProcess != nil && serverProcess.Process != nil {
		serverProcess.Process.Kill()
		time.Sleep(1 * time.Second)
	}
	
	serverProcess = exec.Command("./test-server")
	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr
	if err := serverProcess.Start(); err != nil {
		t.Fatalf("Failed to start server with errors enabled: %v", err)
	}
	
	// Wait for the server to start
	time.Sleep(2 * time.Second)
	
	// Create a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// Make multiple requests to check if any fail
	errorCount := 0
	successCount := 0
	totalRequests := 10
	
	for i := 0; i < totalRequests; i++ {
		resp, err := client.Get(serverURL + "/api/data")
		if err != nil {
			errorCount++
			continue
		}
		
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			errorCount++
		} else {
			successCount++
		}
	}
	
	// We should see some errors with an 80% error rate
	t.Logf("Got %d errors out of %d requests", errorCount, totalRequests)
	if errorCount == 0 {
		t.Error("Expected some errors with high error rate, but got none")
	}
	
	// Restore previous settings and restart server
	os.Setenv("ERROR_RATE", prevErrorRate)
	os.Setenv("SIMULATE_ERRORS", "false")
	
	if serverProcess != nil && serverProcess.Process != nil {
		serverProcess.Process.Kill()
		time.Sleep(1 * time.Second)
	}
	
	serverProcess = exec.Command("./test-server")
	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr
	if err := serverProcess.Start(); err != nil {
		t.Fatalf("Failed to restart server with normal settings: %v", err)
	}
	
	// Wait for the server to restart
	time.Sleep(2 * time.Second)
}
