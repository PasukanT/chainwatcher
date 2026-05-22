package api

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket implements a simple token bucket rate limiter per client IP.
type tokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	maxToken float64
	rate     float64 // tokens per second
	lastFill time.Time
}

var (
	clientBuckets = make(map[string]*tokenBucket)
	bucketsMu     sync.Mutex
)

// RateLimit returns Gin middleware that limits requests per client IP.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		bucketsMu.Lock()
		bucket, exists := clientBuckets[ip]
		if !exists {
			bucket = &tokenBucket{
				tokens:   float64(burst),
				maxToken: float64(burst),
				rate:     rps,
				lastFill: time.Now(),
			}
			clientBuckets[ip] = bucket
		}
		bucketsMu.Unlock()

		bucket.mu.Lock()
		elapsed := time.Since(bucket.lastFill).Seconds()
		bucket.tokens += elapsed * bucket.rate
		if bucket.tokens > bucket.maxToken {
			bucket.tokens = bucket.maxToken
		}
		bucket.lastFill = time.Now()

		if bucket.tokens < 1 {
			bucket.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":    false,
				"error":      "Rate limit exceeded. Please slow down.",
				"retry_after": "1s",
			})
			c.Abort()
			return
		}
		bucket.tokens--
		bucket.mu.Unlock()

		c.Next()
	}
}

// APIKeyAuth returns Gin middleware that validates the X-API-Key header.
func APIKeyAuth(validKeys []string) gin.HandlerFunc {
	keySet := make(map[string]bool, len(validKeys))
	for _, k := range validKeys {
		keySet[strings.TrimSpace(k)] = true
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if !keySet[apiKey] {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or missing API key. Pass via X-API-Key header or api_key query param.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestLogger returns Gin middleware that logs each request with timing.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[CHAINWATCHER] %s %s?%s | %d | %v | %s",
			c.Request.Method, path, query, status, latency, c.ClientIP())
	}
}
