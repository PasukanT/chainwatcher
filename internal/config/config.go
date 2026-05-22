package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Port            int
	RPCURL          string
	APIKeys         []string
	RateLimitRPS    float64
	RateLimitBurst  int
	WhaleThreshold  float64
	AccumWindow     int // seconds
	TVLChangePct    float64
	BadDebtRatio    float64
	GasMultiplier   float64
	HoneypotTaxRate float64
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:            getEnvInt("CW_PORT", 8080),
		RPCURL:          getEnvStr("CW_RPC_URL", "https://eth-mainnet.g.alchemy.com/v2/demo"),
		APIKeys:         getEnvSlice("CW_API_KEYS", []string{"demo-key-001"}),
		RateLimitRPS:    getEnvFloat("CW_RATE_LIMIT_RPS", 10.0),
		RateLimitBurst:  getEnvInt("CW_RATE_LIMIT_BURST", 20),
		WhaleThreshold:  getEnvFloat("CW_WHALE_THRESHOLD", 100000.0),
		AccumWindow:     getEnvInt("CW_ACCUM_WINDOW", 3600),
		TVLChangePct:    getEnvFloat("CW_TVL_CHANGE_PCT", 10.0),
		BadDebtRatio:    getEnvFloat("CW_BAD_DEBT_RATIO", 0.05),
		GasMultiplier:   getEnvFloat("CW_GAS_MULTIPLIER", 1.5),
		HoneypotTaxRate: getEnvFloat("CW_HONEYPOT_TAX", 10.0),
	}
}

func getEnvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}
