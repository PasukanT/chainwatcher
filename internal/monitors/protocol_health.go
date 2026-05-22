package monitors

import (
	"math"
	"sync"
	"time"

	"chainwatcher/internal/models"
)

// protocolData stores historical protocol metrics for trend analysis.
type protocolData struct {
	Name       string
	TVLCurrent float64
	TVLPrev    float64
	Deposits   float64
	Borrows    float64
	BadDebt    float64
	Timestamp  time.Time
}

var (
	phMu        sync.RWMutex
	phProtocols = make(map[string]*protocolData)
)

// CheckProtocolHealth evaluates DeFi protocol health metrics including
// TVL changes, utilization rates, and bad debt ratios.
func CheckProtocolHealth() *models.APIResponse {
	phMu.RLock()
	defer phMu.RUnlock()

	now := time.Now()
	protocols := fetchProtocolData()

	var metrics []models.ProtocolMetrics
	for _, p := range protocols {
		m := evaluateProtocol(p, now)
		metrics = append(metrics, m)
	}

	return &models.APIResponse{
		Success:   true,
		Data:      metrics,
		Count:     len(metrics),
		Timestamp: now,
	}
}

// evaluateProtocol computes health metrics and generates alerts for a single protocol.
func evaluateProtocol(p *protocolData, now time.Time) models.ProtocolMetrics {
	// Calculate TVL change percentage
	tvlChange := 0.0
	if p.TVLPrev > 0 {
		tvlChange = ((p.TVLCurrent - p.TVLPrev) / p.TVLPrev) * 100
	}

	// Calculate utilization ratio (borrows / deposits)
	utilization := 0.0
	if p.Deposits > 0 {
		utilization = (p.Borrows / p.Deposits) * 100
	}

	// Calculate bad debt ratio
	badDebtRatio := 0.0
	if p.Borrows > 0 {
		badDebtRatio = p.BadDebt / p.Borrows
	}

	// Generate health score (100 = perfect health)
	healthScore := computeHealthScore(tvlChange, utilization, badDebtRatio)

	// Generate alerts based on thresholds
	alerts := generateProtocolAlerts(p.Name, tvlChange, utilization, badDebtRatio)

	return models.ProtocolMetrics{
		Protocol:     p.Name,
		TVL:          math.Round(p.TVLCurrent*100) / 100,
		TVLChange24h: math.Round(tvlChange*100) / 100,
		Utilization:  math.Round(utilization*100) / 100,
		BadDebt:      math.Round(badDebtRatio*10000) / 10000,
		HealthScore:  math.Round(healthScore*100) / 100,
		Timestamp:    now,
		Alerts:       alerts,
	}
}

// computeHealthScore generates a composite health score from protocol metrics.
// TVL decline penalizes score, high utilization and bad debt reduce it further.
func computeHealthScore(tvlChange, utilization, badDebtRatio float64) float64 {
	score := 100.0

	// TVL decline penalty (up to -40 points)
	if tvlChange < 0 {
		score += math.Max(tvlChange, -40)
	}
	// TVL growth bonus (up to +10 points)
	if tvlChange > 0 {
		score += math.Min(tvlChange, 10)
	}

	// Utilization penalty: optimal range is 60-80%
	if utilization > 90 {
		score -= (utilization - 90) * 2 // Severe penalty for over-utilization
	} else if utilization > 80 {
		score -= (utilization - 80) * 0.5
	}
	if utilization < 20 {
		score -= (20 - utilization) * 0.3 // Under-utilization is less severe
	}

	// Bad debt penalty (proportional, up to -50 points)
	if badDebtRatio > 0 {
		score -= math.Min(badDebtRatio*1000, 50)
	}

	return math.Max(0, math.Min(100, score))
}

// generateProtocolAlerts creates alert strings for concerning metrics.
func generateProtocolAlerts(name string, tvlChange, utilization, badDebt float64) []string {
	var alerts []string

	if tvlChange < -10 {
		alerts = append(alerts, "TVL_DROP_CRITICAL: TVL dropped >10% in 24h")
	} else if tvlChange < -5 {
		alerts = append(alerts, "TVL_DROP_WARNING: TVL dropped >5% in 24h")
	}

	if utilization > 95 {
		alerts = append(alerts, "UTILIZATION_CRITICAL: Near 100% utilization, withdrawal risk")
	} else if utilization > 85 {
		alerts = append(alerts, "UTILIZATION_WARNING: High utilization >85%")
	}

	if badDebt > 0.05 {
		alerts = append(alerts, "BAD_DEBT_CRITICAL: Bad debt ratio exceeds 5%")
	} else if badDebt > 0.02 {
		alerts = append(alerts, "BAD_DEBT_WARNING: Bad debt ratio exceeds 2%")
	}

	return alerts
}

// fetchProtocolData simulates fetching on-chain protocol data.
func fetchProtocolData() []*protocolData {
	return []*protocolData{
		{Name: "Aave V3", TVLCurrent: 12_500_000_000, TVLPrev: 13_200_000_000, Deposits: 8_200_000_000, Borrows: 4_800_000_000, BadDebt: 1_200_000},
		{Name: "Compound V3", TVLCurrent: 3_800_000_000, TVLPrev: 3_600_000_000, Deposits: 2_500_000_000, Borrows: 1_400_000_000, BadDebt: 800_000},
		{Name: "MakerDAO", TVLCurrent: 8_100_000_000, TVLPrev: 8_300_000_000, Deposits: 6_000_000_000, Borrows: 3_200_000_000, BadDebt: 500_000},
		{Name: "Lido", TVLCurrent: 28_000_000_000, TVLPrev: 27_500_000_000, Deposits: 28_000_000_000, Borrows: 0, BadDebt: 0},
		{Name: "Curve", TVLCurrent: 2_100_000_000, TVLPrev: 2_400_000_000, Deposits: 1_800_000_000, Borrows: 300_000_000, BadDebt: 45_000_000},
	}
}
