package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is process configuration loaded from the environment.
type Config struct {
	DatabaseURL        string
	HTTPAddr           string
	CORSOrigins        []string
	EnableDevEndpoints bool
	CollectorInterval  time.Duration
	LogLevel           string
	DimriaAPIKey       string
	EnableOLX          bool
	ChromePath         string
	MarketCities       []string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:        env("DATABASE_URL", "postgres://estate:estate@localhost:55433/estate?sslmode=disable"),
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		CORSOrigins:        splitCSV(env("CORS_ORIGINS", "http://localhost:5173")),
		EnableDevEndpoints: envBool("ENABLE_DEV_ENDPOINTS", false),
		LogLevel:           env("LOG_LEVEL", "info"),
		DimriaAPIKey:       env("DIMRIA_API_KEY", ""),
		EnableOLX:          envBool("ENABLE_OLX", true),
		ChromePath:         env("CHROME_PATH", ""),
		MarketCities:       splitCSV(env("MARKET_CITIES", "kyiv")),
	}

	interval := env("COLLECTOR_INTERVAL", "2m")
	d, err := time.ParseDuration(interval)
	if err != nil {
		return Config{}, fmt.Errorf("COLLECTOR_INTERVAL: %w", err)
	}
	cfg.CollectorInterval = d
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
