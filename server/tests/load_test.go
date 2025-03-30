package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Unic-X/slow-server/models"
)

// This test file contains load tests to simulate multiple concurrent requests

// TestConcurrentRequests simulates multiple concurrent requests to test load handling
func TestConcurrentRequests(t *testing.T) {
	// Ensure server is running (skip in unit test mode)
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping load test in unit test mode")
	}
	waitForServer(t)

	// Configure the load test
	concurrency := 10  // Number of concurrent users
	requestsPerUser := 5 // Number of requests per user
	totalRequests := concurrency * requestsPerUser

	// Create a channel to collect results
	resultChan := make(chan struct {
		duration time.Duration
		status   int
		err      error
	}, totalRequests)

	// Create a wait group for all goroutines
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Start concurrent users
	for i := 0; i < concurrency; i++ {
		go func(userID int) {
			defer wg.Done()

			// Create a client with timeout for this user
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			// Each user makes multiple requests
			for j := 0; j < requestsPerUser; j++ {
				// Choose an endpoint (alternating between data and users)
				endpoint := "/api/data"
				if j%2 == 1 {
					endpoint = "/api/users"
				}

				// Measure request time
				start := time.Now()
				resp, err := client.Get(serverURL + endpoint)
				duration := time.Since(start)

				// Record the result
				status := 0
				if err == nil {
					status = resp.StatusCode
					resp.Body.Close()
				}

				resultChan <- struct {
					duration time.Duration
					status   int
					err      error
				}{duration, status, err}

				// Small delay between requests from the same user
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// Wait in a goroutine and close the channel when done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect and analyze results
	var successCount, errorCount int
	var totalDuration time.Duration
	var minDuration = time.Hour
	var maxDuration time.Duration
	durations := make([]time.Duration, 0, totalRequests)

	// Process all results from the channel
	for result := range resultChan {
		durations = append(durations, result.duration)
		totalDuration += result.duration

		if result.err != nil {
			errorCount++
			t.Logf("Request error: %v", result.err)
		} else if result.status >= 400 {
			errorCount++
			t.Logf("Request failed with status: %d", result.status)
		} else {
			successCount++
		}

		// Track min/max durations
		if result.duration < minDuration {
			minDuration = result.duration
		}
		if result.duration > maxDuration {
			maxDuration = result.duration
		}
	}

	// Calculate statistics
	avgDuration := totalDuration / time.Duration(len(durations))
	
	// Find p95 (95th percentile) duration
	// First sort the durations (simple bubble sort for small datasets)
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
	p95Index := int(float64(len(durations)) * 0.95)
	p95Duration := durations[p95Index]

	// Log the results
	t.Logf("Load test complete with %d concurrent users, %d requests per user", concurrency, requestsPerUser)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
	t.Logf("Min duration: %v", minDuration)
	t.Logf("Max duration: %v", maxDuration)
	t.Logf("Avg duration: %v", avgDuration)
	t.Logf("P95 duration: %v", p95Duration)

	// Verify that we got some successful responses
	if successCount == 0 {
		t.Error("No successful requests")
	}
}

// TestEndpointThroughput measures how many requests the server can handle over time
func TestEndpointThroughput(t *testing.T) {
	// Ensure server is running (skip in unit test mode)
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping throughput test in unit test mode")
	}
	waitForServer(t)

	// Test duration and concurrency
	testDuration := 10 * time.Second
	concurrency := 5

	// Create a context with timeout for the test duration
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	// Channel for collecting results
	resultChan := make(chan bool, 1000) // true = success, false = error

	// Start concurrent workers
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			// Keep making requests until context is done
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Make a request to the data endpoint (fast endpoint for throughput test)
					resp, err := client.Get(serverURL + "/api/data")
					if err != nil {
						resultChan <- false
						continue
					}

					// Check if response is valid
					var dataResp models.DataResponse
					err = json.NewDecoder(resp.Body).Decode(&dataResp)
					resp.Body.Close()

					if err != nil || resp.StatusCode != http.StatusOK {
						resultChan <- false
					} else {
						resultChan <- true
					}
				}
			}
		}()
	}

	// Wait for test duration to complete
	<-ctx.Done()

	// Signal workers to stop
	cancel()

	// Close result channel after all workers finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	successCount := 0
	errorCount := 0
	for result := range resultChan {
		if result {
			successCount++
		} else {
			errorCount++
		}
	}

	// Calculate throughput
	totalRequests := successCount + errorCount
	requestsPerSecond := float64(totalRequests) / testDuration.Seconds()
	successRate := float64(successCount) / float64(totalRequests) * 100

	// Log results
	t.Logf("Throughput test completed")
	t.Logf("Total requests: %d", totalRequests)
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
	t.Logf("Throughput: %.2f requests/second", requestsPerSecond)
	t.Logf("Success rate: %.2f%%", successRate)

	// Verify that we got some throughput
	if totalRequests == 0 {
		t.Error("No requests were processed during throughput test")
	}
}

// TestEndpointLatencyUnderLoad measures endpoint latency with increasing load
func TestEndpointLatencyUnderLoad(t *testing.T) {
	// Ensure server is running (skip in unit test mode)
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping latency test in unit test mode")
	}
	waitForServer(t)

	// Test increasing concurrency levels
	concurrencyLevels := []int{1, 3, 5, 10}
	endpoint := "/api/data"
	requestsPerLevel := 20

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency%d", concurrency), func(t *testing.T) {
			resultChan := make(chan time.Duration, requestsPerLevel)
			var wg sync.WaitGroup
			wg.Add(concurrency)

			// Start concurrent workers
			for i := 0; i < concurrency; i++ {
				go func(workerID int) {
					defer wg.Done()
					client := &http.Client{
						Timeout: 10 * time.Second,
					}

					requestsPerWorker := requestsPerLevel / concurrency
					if requestsPerWorker < 1 {
						requestsPerWorker = 1
					}

					for j := 0; j < requestsPerWorker; j++ {
						start := time.Now()
						resp, err := client.Get(serverURL + endpoint)
						if err != nil {
							t.Logf("Worker %d request error: %v", workerID, err)
							continue
						}
						duration := time.Since(start)
						resp.Body.Close()

						if resp.StatusCode == http.StatusOK {
							resultChan <- duration
						}

						// Small pause between requests
						time.Sleep(50 * time.Millisecond)
					}
				}(i)
			}

			// Close channel after all workers finish
			go func() {
				wg.Wait()
				close(resultChan)
			}()

			// Collect and analyze durations
			var durations []time.Duration
			for duration := range resultChan {
				durations = append(durations, duration)
			}

			// Skip if no successful requests
			if len(durations) == 0 {
				t.Logf("No successful requests at concurrency level %d", concurrency)
				return
			}

			// Calculate statistics
			var totalDuration time.Duration
			var minDuration = time.Hour
			var maxDuration time.Duration

			for _, d := range durations {
				totalDuration += d
				if d < minDuration {
					minDuration = d
				}
				if d > maxDuration {
					maxDuration = d
				}
			}

			avgDuration := totalDuration / time.Duration(len(durations))

			// Sort durations for percentiles
			for i := 0; i < len(durations); i++ {
				for j := i + 1; j < len(durations); j++ {
					if durations[i] > durations[j] {
						durations[i], durations[j] = durations[j], durations[i]
					}
				}
			}

			// Calculate p50 and p95
			p50Index := int(float64(len(durations)) * 0.5)
			p95Index := int(float64(len(durations)) * 0.95)
			if p95Index >= len(durations) {
				p95Index = len(durations) - 1
			}

			p50Duration := durations[p50Index]
			p95Duration := durations[p95Index]

			// Log results
			t.Logf("Latency at concurrency level %d:", concurrency)
			t.Logf("  Requests: %d", len(durations))
			t.Logf("  Min: %v", minDuration)
			t.Logf("  Max: %v", maxDuration)
			t.Logf("  Avg: %v", avgDuration)
			t.Logf("  p50: %v", p50Duration)
			t.Logf("  p95: %v", p95Duration)

			// Verify that latency is within reasonable bounds
			if avgDuration < 100*time.Millisecond {
				t.Errorf("Average latency too low for slow server: %v", avgDuration)
			}
		})
	}
}
