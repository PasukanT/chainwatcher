package monitors

import (
	"math"
	"sort"
	"sync"
	"time"

	"chainwatcher/internal/models"
)

// transferRecord stores historical transfer data for pattern detection.
type transferRecord struct {
	Address   string
	AmountUSD float64
	Timestamp time.Time
	IsBuy     bool
}

var (
	whaleMu       sync.RWMutex
	whaleHistory  = make(map[string][]transferRecord) // address -> transfers
	whaleAlerts   []models.WhaleAlert
)

// ScanWhaleTransfers analyzes recent transfers and detects whale activity.
// It flags transfers above the given threshold and identifies accumulation
// patterns (3+ buys within the analysis window).
func ScanWhaleTransfers(thresholdUSD float64, windowSec int) *models.APIResponse {
	whaleMu.RLock()
	defer whaleMu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)

	// Simulate scanning recent blocks for large transfers
	transfers := fetchRecentTransfers(thresholdUSD, cutoff)

	var alerts []models.WhaleAlert
	for _, t := range transfers {
		alert := classifyTransfer(t, thresholdUSD, windowSec)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Detect accumulation patterns: 3+ buys in window
	accumAlerts := detectAccumulation(transfers, cutoff, 3)
	alerts = append(alerts, accumAlerts...)

	// Detect distribution patterns: 3+ sells in window
	distAlerts := detectDistribution(transfers, cutoff, 3)
	alerts = append(alerts, distAlerts...)

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].ValueUSD > alerts[j].ValueUSD
	})

	return &models.APIResponse{
		Success:   true,
		Data:      alerts,
		Count:     len(alerts),
		Timestamp: now,
	}
}

// classifyTransfer assigns an alert type based on transfer characteristics.
func classifyTransfer(t transferRecord, threshold float64, windowSec int) *models.WhaleAlert {
	if t.AmountUSD < threshold {
		return nil
	}

	alertType := "whale_transfer"
	direction := "transfer"
	if t.IsBuy {
		alertType = "whale_buy"
		direction = "accumulation"
	} else {
		alertType = "whale_sell"
		direction = "distribution"
	}

	return &models.WhaleAlert{
		TxHash:    generateTxHash(t.Address, t.Timestamp),
		From:      t.Address,
		To:        "0x" + t.Address[:8] + "...",
		ValueUSD:  t.AmountUSD,
		Token:     "ETH",
		Direction: direction,
		Timestamp: t.Timestamp,
		AlertType: alertType,
	}
}

// detectAccumulation finds addresses with >= minBuys purchases within the window.
func detectAccumulation(transfers []transferRecord, cutoff time.Time, minBuys int) []models.WhaleAlert {
	buyCounts := make(map[string]int)
	buyTotal := make(map[string]float64)

	for _, t := range transfers {
		if t.IsBuy && t.Timestamp.After(cutoff) {
			buyCounts[t.Address]++
			buyTotal[t.Address] += t.AmountUSD
		}
	}

	var alerts []models.WhaleAlert
	for addr, count := range buyCounts {
		if count >= minBuys {
			alerts = append(alerts, models.WhaleAlert{
				TxHash:    "accum_" + addr[:8],
				From:      addr,
				ValueUSD:  buyTotal[addr],
				Token:     "ETH",
				Direction: "accumulation",
				Timestamp: time.Now(),
				AlertType: "accumulation_pattern",
			})
		}
	}
	return alerts
}

// detectDistribution finds addresses with >= minSells sells within the window.
func detectDistribution(transfers []transferRecord, cutoff time.Time, minSells int) []models.WhaleAlert {
	sellCounts := make(map[string]int)
	sellTotal := make(map[string]float64)

	for _, t := range transfers {
		if !t.IsBuy && t.Timestamp.After(cutoff) {
			sellCounts[t.Address]++
			sellTotal[t.Address] += t.AmountUSD
		}
	}

	var alerts []models.WhaleAlert
	for addr, count := range sellCounts {
		if count >= minSells {
			alerts = append(alerts, models.WhaleAlert{
				TxHash:    "dist_" + addr[:8],
				From:      addr,
				ValueUSD:  sellTotal[addr],
				Token:     "ETH",
				Direction: "distribution",
				Timestamp: time.Now(),
				AlertType: "distribution_pattern",
			})
		}
	}
	return alerts
}

// fetchRecentTransfers simulates fetching on-chain transfer data.
// In production, this would query an RPC node or indexer API.
func fetchRecentTransfers(threshold float64, since time.Time) []transferRecord {
	now := time.Now()
	addrs := []string{
		"d8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
		"Ab5801a7D398351b8bE11C439e05C5B3259aeC9B",
		"fB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"28C6c06298d514Db089934071355E5743bf21d60",
		"21a31Ee1afC51d94C2eFcCAa2092aD1028285549",
	}

	var records []transferRecord
	for i, addr := range addrs {
		amount := 50000.0 + float64(i)*120000.0
		records = append(records, transferRecord{
			Address:   addr,
			AmountUSD: math.Round(amount*100) / 100,
			Timestamp: now.Add(-time.Duration(i*300) * time.Second),
			IsBuy:     i%3 != 0,
		})
	}
	return records
}

func generateTxHash(addr string, ts time.Time) string {
	return "0x" + addr[:16] + ts.Format("060102150405")
}
