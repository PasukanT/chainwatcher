package monitors

import (
	"math"
	"sort"
	"sync"
	"time"

	"chainwatcher/internal/models"
)

// pendingTx represents a raw pending transaction from the mempool.
type pendingTx struct {
	Hash      string
	From      string
	To        string
	GasPrice  float64 // in gwei
	GasTipCap float64
	Value     float64 // in ETH
	Data      []byte
	Timestamp time.Time
}

var (
	mpMu       sync.RWMutex
	mpPending  []pendingTx
)

// AnalyzeMempool monitors pending transactions for MEV activity,
// sandwich attacks, and frontrunning patterns.
func AnalyzeMempool(windowSec int) *models.APIResponse {
	mpMu.RLock()
	defer mpMu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)

	// Fetch pending transactions from mempool
	pending := fetchPendingTxs(cutoff)

	// Filter to window
	var filtered []pendingTx
	for _, tx := range pending {
		if tx.Timestamp.After(cutoff) {
			filtered = append(filtered, tx)
		}
	}

	// Detect sandwich attacks
	sandwichFlags := detectSandwichAttacks(filtered)

	// Detect frontrunning
	frontrunFlags := detectFrontrunning(filtered)

	// Analyze gas price anomalies
	gasFlags := analyzeGasAnomalies(filtered)

	// Build response with risk assessments
	var results []models.MempoolTx
	for _, tx := range filtered {
		mt := classifyMempoolTx(tx, sandwichFlags, frontrunFlags, gasFlags)
		results = append(results, mt)
	}

	sort.Slice(results, func(i, j int) bool {
		return riskLevel(results[i].Risk) > riskLevel(results[j].Risk)
	})

	return &models.APIResponse{
		Success:   true,
		Data:      results,
		Count:     len(results),
		Timestamp: now,
	}
}

// classifyMempoolTx assigns flags and risk levels to a pending transaction.
func classifyMempoolTx(tx pendingTx, sandwich, frontrun, gas map[string][]string) models.MempoolTx {
	var flags []string
	risk := "low"

	if sf, ok := sandwich[tx.Hash]; ok {
		flags = append(flags, sf...)
		risk = "critical"
	}
	if ff, ok := frontrun[tx.Hash]; ok {
		flags = append(flags, ff...)
		if risk != "critical" {
			risk = "high"
		}
	}
	if gf, ok := gas[tx.Hash]; ok {
		flags = append(flags, gf...)
		if risk == "low" {
			risk = "medium"
		}
	}

	return models.MempoolTx{
		TxHash:      tx.Hash,
		From:        tx.From,
		To:          tx.To,
		GasPrice:    math.Round(tx.GasPrice*100) / 100,
		GasTipCap:   math.Round(tx.GasTipCap*100) / 100,
		Value:       math.Round(tx.Value*1000) / 1000,
		MethodSig:   "0x38ed1739", // swapExactTokensForTokens
		PendingTime: math.Round(time.Since(tx.Timestamp).Seconds()*10) / 10,
		Flags:       flags,
		Risk:        risk,
	}
}

// detectSandwichAttacks identifies transactions that form sandwich patterns:
// a frontrun tx before a victim tx and a backrun tx after.
func detectSandwichAttacks(txs []pendingTx) map[string][]string {
	results := make(map[string][]string)

	// Sort by gas price descending (higher gas = likely frontrun)
	sorted := make([]pendingTx, len(txs))
	copy(sorted, txs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GasPrice > sorted[j].GasPrice
	})

	// Look for pairs where same sender has a buy then sell around a victim tx
	for i := 0; i < len(sorted)-2; i++ {
		for j := i + 1; j < len(sorted)-1; j++ {
			if sorted[i].From == sorted[j+1].From && sorted[i].To == sorted[j+1].To {
				// Same sender, same contract, sandwich pattern
				gasRatio := sorted[i].GasPrice / sorted[j].GasPrice
				if gasRatio > 1.5 && sorted[i].GasPrice > 100 {
					results[sorted[i].Hash] = []string{"sandwich_frontrun"}
					results[sorted[j].Hash] = []string{"sandwich_victim"}
					results[sorted[j+1].Hash] = []string{"sandwich_backrun"}
				}
			}
		}
	}
	return results
}

// detectFrontrunning finds transactions with suspiciously higher gas prices
// targeting the same contract as nearby lower-gas transactions.
func detectFrontrunning(txs []pendingTx) map[string][]string {
	results := make(map[string][]string)

	// Group by target contract
	contractTxs := make(map[string][]pendingTx)
	for _, tx := range txs {
		contractTxs[tx.To] = append(contractTxs[tx.To], tx)
	}

	for _, group := range contractTxs {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].GasPrice > group[j].GasPrice
		})
		// If top tx has 2x+ gas of median, flag as frontrun
		median := group[len(group)/2].GasPrice
		if group[0].GasPrice > median*2 && group[0].GasPrice > 50 {
			results[group[0].Hash] = []string{"frontrun"}
		}
	}
	return results
}

// analyzeGasAnomalies flags transactions with abnormally high gas prices.
func analyzeGasAnomalies(txs []pendingTx) map[string][]string {
	results := make(map[string][]string)

	if len(txs) == 0 {
		return results
	}

	// Calculate median gas price
	prices := make([]float64, len(txs))
	for i, tx := range txs {
		prices[i] = tx.GasPrice
	}
	sort.Float64s(prices)
	median := prices[len(prices)/2]

	// Flag transactions with gas > 3x median
	for _, tx := range txs {
		if tx.GasPrice > median*3 && tx.GasPrice > 200 {
			results[tx.Hash] = []string{"gas_anomaly"}
		}
	}
	return results
}

func riskLevel(r string) int {
	switch r {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

// fetchPendingTxs simulates mempool data retrieval.
func fetchPendingTxs(since time.Time) []pendingTx {
	now := time.Now()
	return []pendingTx{
		{Hash: "0xaaaa111122223333444455556666777788889999aaaabbbbccccddddeeeeffff", From: "0xbot1", To: "0xUniswapRouter", GasPrice: 350, GasTipCap: 50, Value: 12.5, Timestamp: now.Add(-10 * time.Second)},
		{Hash: "0xbbbb111122223333444455556666777788889999aaaabbbbccccddddeeeeffff", From: "0xuser1", To: "0xUniswapRouter", GasPrice: 45, GasTipCap: 5, Value: 2.1, Timestamp: now.Add(-30 * time.Second)},
		{Hash: "0xcccc111122223333444455556666777788889999aaaabbbbccccddddeeeeffff", From: "0xbot1", To: "0xUniswapRouter", GasPrice: 360, GasTipCap: 55, Value: 13.0, Timestamp: now.Add(-5 * time.Second)},
		{Hash: "0xdddd111122223333444455556666777788889999aaaabbbbccccddddeeeeffff", From: "0xmev2", To: "0xSushiRouter", GasPrice: 200, GasTipCap: 80, Value: 50.0, Timestamp: now.Add(-15 * time.Second)},
		{Hash: "0xeeee111122223333444455556666777788889999aaaabbbbccccddddeeeeffff", From: "0xuser2", To: "0xSushiRouter", GasPrice: 30, GasTipCap: 3, Value: 0.5, Timestamp: now.Add(-45 * time.Second)},
	}
}
