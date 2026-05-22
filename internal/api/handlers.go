package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"chainwatcher/internal/monitors"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// GetWhaleAlerts returns whale transfer alerts above threshold.
// GET /api/whales?threshold=100000&window=3600
func GetWhaleAlerts(c *gin.Context) {
	threshold := parseFloatParam(c, "threshold", 100000)
	window := parseIntParam(c, "window", 3600)

	result := monitors.ScanWhaleTransfers(threshold, window)
	c.JSON(http.StatusOK, result)
}

// GetSmartMoneySignals returns detected smart money wallet signals.
// GET /api/smart-money?window=3600&min_score=60
func GetSmartMoneySignals(c *gin.Context) {
	window := parseIntParam(c, "window", 3600)

	result := monitors.AnalyzeSmartMoney(window)
	c.JSON(http.StatusOK, result)
}

// GetProtocolHealth returns DeFi protocol health metrics.
// GET /api/health
func GetProtocolHealth(c *gin.Context) {
	result := monitors.CheckProtocolHealth()
	c.JSON(http.StatusOK, result)
}

// GetMempoolActivity returns pending mempool transaction analysis.
// GET /api/mempool?window=300
func GetMempoolActivity(c *gin.Context) {
	window := parseIntParam(c, "window", 300)

	result := monitors.AnalyzeMempool(window)
	c.JSON(http.StatusOK, result)
}

// GetTokenAlerts returns newly deployed token risk assessments.
// GET /api/tokens?window=3600
func GetTokenAlerts(c *gin.Context) {
	window := parseIntParam(c, "window", 3600)

	result := monitors.ScanNewTokens(window)
	c.JSON(http.StatusOK, result)
}

// HealthCheck is a public endpoint for load balancer health checks.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
		"uptime":  fmt.Sprintf("%.0fs", time.Since(startTime).Seconds()),
	})
}

// parseFloatParam extracts a float query parameter with a default value.
func parseFloatParam(c *gin.Context, key string, defaultVal float64) float64 {
	val := c.DefaultQuery(key, "")
	if val == "" {
		return defaultVal
	}
	result, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return result
}

// parseIntParam extracts an integer query parameter with a default value.
func parseIntParam(c *gin.Context, key string, defaultVal int) int {
	val := c.DefaultQuery(key, "")
	if val == "" {
		return defaultVal
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}
