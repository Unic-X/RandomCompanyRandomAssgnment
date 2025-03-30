package config

import (
	"github.com/charmbracelet/log"
	"os"
	"strconv"
)

type Config struct {
	Port           int
	LogLevel       string
	SimulateErrors bool
	MinDelay      int
	MaxDelay      int
	DBQueryDelay  int
	APICallDelay  int
	ProcessDelay  int
	ErrorRate     float64 // Percentage of errors 0.00 = 0% error and 1.00 = 100% error
	EnableMetrics bool
}

func LoadConfig() *Config {
	cfg := &Config{
		Port:           8080,
		LogLevel:       "info",
		SimulateErrors: true,
		MinDelay:       500,  
		MaxDelay:       3000, 
		DBQueryDelay:   800,  // 800ms for DB query simulation
		APICallDelay:   1200, // 1.2s for external API call simulation
		ProcessDelay:   500,  // 500ms for processing simulation
		ErrorRate:      0.15, // 15% error rate
		EnableMetrics:  true,
	}

	if port := os.Getenv("SERVER_PORT"); port != "" { //Hardcoded inside Dockerfile for now
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		} else {
			log.Printf("Invalid port: %s, using default: %d", port, cfg.Port)
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if simErr := os.Getenv("SIMULATE_ERRORS"); simErr == "false" {
		cfg.SimulateErrors = false
	}

	if errRate := os.Getenv("ERROR_RATE"); errRate != "" {
		if er, err := strconv.ParseFloat(errRate, 64); err == nil && er >= 0 && er <= 1 {
			cfg.ErrorRate = er
		}
	}

	return cfg
}

//Return port as a string
func (c *Config) PortString() string {
	return strconv.Itoa(c.Port)
}
