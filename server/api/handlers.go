package api

import (
	"encoding/json"
	"github.com/charmbracelet/log"
	"math/rand"
	"net/http"
	"github.com/Unic-X/slow-server/config"
	"github.com/Unic-X/slow-server/metrics"
	"github.com/Unic-X/slow-server/models"
	"time"
)

var cfg *config.Config

func init() {
	cfg = config.LoadConfig()
	rand.NewSource(time.Now().UnixNano())
}

func SetConfig(c *config.Config) {
	cfg = c
}

func simulateDelay(min, max int) {
	delay := min
	if max > min {
		delay = min + rand.Intn(max-min)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func simulateError() bool {
	if !cfg.SimulateErrors {
		return false
	}
	return rand.Float64() < cfg.ErrorRate
}

func simulateDBQuery() (bool, error) {
	startTime := time.Now()
	simulateDelay(cfg.DBQueryDelay/2, cfg.DBQueryDelay)
	duration := time.Since(startTime)
	
	metrics.DBQueryDuration.Observe(float64(duration.Milliseconds()))
	metrics.DBQueriesTotal.Inc()
	
	if simulateError() {
		log.Errorf("Database query failed after %v", duration)
		metrics.DBQueryErrors.Inc()
		return false, models.NewAppError("Database query failed", http.StatusInternalServerError)
	}
	
	return true, nil
}

func simulateExternalAPICall() (bool, error) {
	startTime := time.Now()
	simulateDelay(cfg.APICallDelay/2, cfg.APICallDelay)
	duration := time.Since(startTime)
	
	metrics.ExternalAPICallDuration.Observe(float64(duration.Milliseconds()))
	metrics.ExternalAPICallsTotal.Inc()
	
	if simulateError() {
		log.Errorf("External API call failed after %v", duration)
		metrics.ExternalAPICallErrors.Inc()
		return false, models.NewAppError("External API call failed", http.StatusBadGateway)
	}
	
	return true, nil
}

func simulateProcessing() (bool, error) {
	startTime := time.Now()
	simulateDelay(cfg.ProcessDelay/2, cfg.ProcessDelay)
	duration := time.Since(startTime)
	
	metrics.ProcessingDuration.Observe(float64(duration.Milliseconds()))
	
	if simulateError() {
		log.Warnf("Processing failed after %v", duration)
		metrics.ProcessingErrors.Inc()
		return false, models.NewAppError("Processing failed", http.StatusInternalServerError)
	}
	
	return true, nil
}

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Header.Get("X-Request-ID")
	
	log.Infof("[%s] Processing GET /api/data request", requestID)
	
	_, err := simulateDBQuery()
	if err != nil {
		log.Infof("[%s] Error in DB query: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	_, err = simulateProcessing()
	if err != nil {
		log.Errorf("[%s] Error in processing: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := models.DataResponse{ //Sample Response
		Data: []models.DataItem{
			{ID: 1, Name: "Item 1", Value: rand.Float64() * 100},
			{ID: 2, Name: "Item 2", Value: rand.Float64() * 100},
			{ID: 3, Name: "Item 3", Value: rand.Float64() * 100},
		},
		Count: 3,
	}
	
	duration := time.Since(startTime)
	log.Infof("[%s] GET /api/data completed in %v", requestID, duration)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Header.Get("X-Request-ID")
	
	log.Infof("[%s] Processing GET /api/users request", requestID)
	
	_, err := simulateDBQuery()
	if err != nil {
		log.Errorf("[%s] Error in DB query: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	_, err = simulateExternalAPICall()
	if err != nil {
		log.Errorf("[%s] Error in external API call: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	
	response := models.UsersResponse{ //Sample response
		Users: []models.User{
			{ID: 1, Name: "User 1", Email: "user1@example.com"},
			{ID: 2, Name: "User 2", Email: "user2@example.com"},
			{ID: 3, Name: "User 3", Email: "user3@example.com"},
		},
		Count: 3,
	}
	
	duration := time.Since(startTime)
	log.Infof("[%s] GET /api/users completed in %v", requestID, duration)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ProcessDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	startTime := time.Now()
	requestID := r.Header.Get("X-Request-ID")
	
	log.Infof("[%s] Processing POST /api/process request", requestID)
	
	var request models.ProcessRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Errorf("[%s] Error parsing request body: %v", requestID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	_, err = simulateDBQuery()
	if err != nil {
		log.Errorf("[%s] Error in DB query: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	_, err = simulateProcessing()
	if err != nil {
		log.Errorf("[%s] Error in processing: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	_, err = simulateProcessing()
	if err != nil {
		log.Errorf("[%s] Error in secondary processing: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	_, err = simulateExternalAPICall()
	if err != nil {
		log.Errorf("[%s] Error in external API call: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	
	response := models.ProcessResponse{
		Success:   true,
		ProcessID: rand.Intn(10000),
		Message:   "Data processed successfully",
	}
	
	duration := time.Since(startTime)
	log.Printf("[%s] POST /api/process completed in %v", requestID, duration)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
