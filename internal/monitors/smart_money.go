package monitors

import (
	"math"
	"sort"
	"sync"
	"time"

	"chainwatcher/internal/models"
)

// walletTrack holds performance metrics for a tracked wallet.
type walletTrack struct {
	Address     string
	Trades      []tradeRecord
	TotalPnL    float64
	MEVCount    int
	LastSeen    time.Time
}

type tradeRecord struct {
	Token     string
	EntryUSD  float64
	ExitUSD   float64
	Duration  time.Duration
	Timestamp time.Time
}

var (
	smMu      sync.RWMutex
	smWallets = make(map[string]*walletTrack)
)

// AnalyzeSmartMoney scans for wallets with consistently high performance,
// calculates win rate, profit factor, and generates composite scores.
func AnalyzeSmartMoney(windowSec int) *models.APIResponse {
	smMu.RLock()
	defer smMu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)

	// Fetch wallet performance data from indexer
	wallets := fetchWalletPerformance(cutoff)

	var signals []models.SmartMoneySignal
	for _, w := range wallets {
		signal := scoreWallet(w)
		if signal != nil && signal.Score >= 60 { // Only report high-confidence signals
			signals = append(signals, *signal)
		}
	}

	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Score > signals[j].Score
	})

	return &models.APIResponse{
		Success:   true,
		Data:      signals,
		Count:     len(signals),
		Timestamp: now,
	}
}

// scoreWallet computes a composite 0-100 score based on multiple factors.
func scoreWallet(w *walletTrack) *models.SmartMoneySignal {
	if len(w.Trades) < 5 {
		return nil // Minimum trade threshold
	}

	wins, totalProfit := 0, 0.0
	totalLoss := 0.0
	for _, t := range w.Trades {
		pnl := t.ExitUSD - t.EntryUSD
		if pnl > 0 {
			wins++
			totalProfit += pnl
		} else {
			totalLoss += math.Abs(pnl)
		}
	}

	winRate := float64(wins) / float64(len(w.Trades))
	profitFactor := 1.0
	if totalLoss > 0 {
		profitFactor = totalProfit / totalLoss
	}
	netPnL := totalProfit - totalLoss

	// Composite score: weighted combination
	// Win rate (30%), profit factor (25%), net PnL (20%), trade count (15%), consistency (10%)
	winScore := math.Min(winRate*100, 100) * 0.30
	pfScore := math.Min(profitFactor*20, 100) * 0.25
	pnlScore := math.Min(netPnL/1000, 100) * 0.20
	countScore := math.Min(float64(len(w.Trades))/50*100, 100) * 0.15
	mevPenalty := 0.0
	if w.MEVCount > len(w.Trades)/2 {
		mevPenalty = 15 // Penalize heavy MEV wallets
	}
	score := winScore + pfScore + pnlScore + countScore - mevPenalty
	score = math.Max(0, math.Min(100, score))

	tags := classifyWallet(w, winRate, profitFactor)

	return &models.SmartMoneySignal{
		Wallet:       w.Address,
		Score:        math.Round(score*100) / 100,
		WinRate:      math.Round(winRate*10000) / 100,
		ProfitFactor: math.Round(profitFactor*100) / 100,
		TotalTrades:  len(w.Trades),
		NetPnL:       math.Round(netPnL*100) / 100,
		Tags:         tags,
		LastActive:   w.LastSeen,
	}
}

// classifyWallet assigns tags based on behavior patterns.
func classifyWallet(w *walletTrack, winRate, pf float64) []string {
	var tags []string
	if w.MEVCount > 0 {
		tags = append(tags, "MEV")
	}
	if winRate > 0.7 && pf > 3.0 {
		tags = append(tags, "copytrade")
	}
	if hasInsiderPattern(w.Trades) {
		tags = append(tags, "insider")
	}
	if winRate > 0.6 && len(w.Trades) > 20 {
		tags = append(tags, "smart_money")
	}
	return tags
}

// hasInsiderPattern detects suspicious pre-pump entry patterns.
func hasInsiderPattern(trades []tradeRecord) bool {
	if len(trades) < 3 {
		return false
	}
	// Look for rapid profitable entries (bought within 1hr of token launch)
	earlyWins := 0
	for _, t := range trades {
		if t.Duration < 2*time.Hour && (t.ExitUSD-t.EntryUSD) > t.EntryUSD*0.5 {
			earlyWins++
		}
	}
	return earlyWins >= 3
}

// fetchWalletPerformance simulates fetching wallet data from an indexer.
func fetchWalletPerformance(since time.Time) []*walletTrack {
	now := time.Now()
	wallets := []*walletTrack{
		{Address: "0xSmartWhale1a2b3c4d5e6f7890abcdef12345678", LastSeen: now},
		{Address: "0xMEVBot9876543210fedcba9876543210fedcba", LastSeen: now},
		{Address: "0xInsiderTraderabcdef1234567890abcdef12", LastSeen: now},
	}

	// Generate realistic trade histories
	for _, w := range wallets {
		for i := 0; i < 25; i++ {
			entry := 1000 + float64(i)*500
			exit := entry * (0.7 + float64(i%5)*0.3)
			w.Trades = append(w.Trades, tradeRecord{
				Token:     "TOKEN" + string(rune('A'+i%26)),
				EntryUSD:  entry,
				ExitUSD:   exit,
				Duration:  time.Duration(i*100) * time.Minute,
				Timestamp: now.Add(-time.Duration(i*3600) * time.Second),
			})
		}
		w.MEVCount = 3 + len(w.Address)%5
	}
	return wallets
}
