package models

import "time"

// WhaleAlert represents a detected whale transaction.
type WhaleAlert struct {
	TxHash      string    `json:"tx_hash"`
	BlockNumber uint64    `json:"block_number"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	ValueUSD    float64   `json:"value_usd"`
	Token       string    `json:"token"`
	Direction   string    `json:"direction"` // "accumulation" | "distribution" | "transfer"
	Timestamp   time.Time `json:"timestamp"`
	AlertType   string    `json:"alert_type"`
}

// SmartMoneySignal represents a profitable wallet detection.
type SmartMoneySignal struct {
	Wallet       string    `json:"wallet"`
	Score        float64   `json:"score"` // 0-100 composite score
	WinRate      float64   `json:"win_rate"`
	ProfitFactor float64   `json:"profit_factor"`
	TotalTrades  int       `json:"total_trades"`
	NetPnL       float64   `json:"net_pnl_usd"`
	Tags         []string  `json:"tags"` // "MEV", "copytrade", "insider"
	LastActive   time.Time `json:"last_active"`
}

// ProtocolMetrics holds DeFi protocol health data.
type ProtocolMetrics struct {
	Protocol     string    `json:"protocol"`
	TVL          float64   `json:"tvl_usd"`
	TVLChange24h float64   `json:"tvl_change_24h_pct"`
	Utilization  float64   `json:"utilization_rate"`
	BadDebt      float64   `json:"bad_debt_ratio"`
	HealthScore  float64   `json:"health_score"` // 0-100
	Timestamp    time.Time `json:"timestamp"`
	Alerts       []string  `json:"alerts,omitempty"`
}

// MempoolTx represents a pending transaction in the mempool.
type MempoolTx struct {
	TxHash      string    `json:"tx_hash"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	GasPrice    float64   `json:"gas_price_gwei"`
	GasTipCap   float64   `json:"gas_tip_cap_gwei"`
	Value       float64   `json:"value_eth"`
	MethodSig   string    `json:"method_signature"`
	PendingTime float64   `json:"pending_seconds"`
	Flags       []string  `json:"flags,omitempty"` // "sandwich", "frontrun", "mev"
	Risk        string    `json:"risk"`             // "low", "medium", "high", "critical"
}

// TokenInfo represents a newly deployed or scanned token.
type TokenInfo struct {
	Address       string    `json:"address"`
	Name          string    `json:"name"`
	Symbol        string    `json:"symbol"`
	Deployer      string    `json:"deployer"`
	DeployBlock   uint64    `json:"deploy_block"`
	LiquidityUSD  float64   `json:"liquidity_usd"`
	BuyTax        float64   `json:"buy_tax_pct"`
	SellTax       float64   `json:"sell_tax_pct"`
	HoneypotFlags []string  `json:"honeypot_flags,omitempty"`
	RiskScore     float64   `json:"risk_score"` // 0-100, higher = riskier
	IsRenounced   bool      `json:"is_renounced"`
	Timestamp     time.Time `json:"timestamp"`
}

// APIResponse wraps any response with metadata.
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Count     int         `json:"count"`
	Timestamp time.Time   `json:"timestamp"`
	Error     string      `json:"error,omitempty"`
}
