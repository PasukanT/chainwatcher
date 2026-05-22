package main

import (
	"fmt"
	"log"
	"os"

	"chainwatcher/internal/api"
	"chainwatcher/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if os.Getenv("CHAINWATCHER_CLI") == "1" {
		runCLI(os.Args[1:])
		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Apply middleware chain
	r.Use(api.RequestLogger())
	r.Use(api.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))

	// Register protected routes
	authorized := r.Group("/api", api.APIKeyAuth(cfg.APIKeys))
	{
		authorized.GET("/whales", api.GetWhaleAlerts)
		authorized.GET("/smart-money", api.GetSmartMoneySignals)
		authorized.GET("/health", api.GetProtocolHealth)
		authorized.GET("/mempool", api.GetMempoolActivity)
		authorized.GET("/tokens", api.GetTokenAlerts)
	}

	// Public endpoints
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": "1.0.0"})
	})
	r.GET("/health", api.HealthCheck)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("ChainWatcher starting on %s", addr)
	log.Printf("Endpoints: /api/whales, /api/smart-money, /api/health, /api/mempool, /api/tokens")
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
