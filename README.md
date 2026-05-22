# ChainWatcher

**Real-time on-chain activity monitor** — whale tracking, smart money alerts, protocol health dashboards, mempool analysis, and new token risk scoring.

Built with Go 1.21+ and the Gin web framework. Designed for speed, composability, and extensibility.

---

## 🔍 Monitor Modules

### 1. Whale Tracker (`/api/whales`)
- Detects large transfers above configurable USD thresholds
- Identifies **accumulation patterns** (3+ buys within 1 hour)
- Identifies **distribution patterns** (3+ sells within 1 hour)
- Real-time alerts for whale wallet movements

### 2. Smart Money (`/api/smart-money`)
- Tracks wallets with high win rates and profit factors
- Composite scoring: win rate (30%), profit factor (25%), net PnL (20%), trade count (15%)
- Tags: `MEV`, `copytrade`, `insider`, `smart_money`
- Filters for 60+ score confidence threshold

### 3. Protocol Health (`/api/health`)
- Monitors TVL changes across major DeFi protocols (Aave, Compound, MakerDAO, Lido, Curve)
- Utilization ratio tracking (optimal: 60-80%)
- Bad debt ratio alerts (warning: 2%, critical: 5%)
- Composite health score 0-100

### 4. Mempool Monitor (`/api/mempool`)
- Pending transaction analysis with gas price anomalies
- **Sandwich attack detection** (frontrun → victim → backrun pattern)
- **Frontrunning alerts** (2x+ gas of median for same-contract targeting)
- Risk classification: low / medium / high / critical

### 5. Token Scanner (`/api/tokens`)
- New token deployment detection and risk scoring
- **Honeypot patterns**: no sell function, high tax, proxy contracts
- **Rug pull red flags**: unlocked liquidity, repeat deployers, non-renounced ownership
- Risk score 0-100 with detailed flag breakdown

---

## 🏗 Architecture

```
main.go                      → Gin server, port 8080, route registration
cli.go                       → CLI entrypoint with flag-based subcommands
internal/
├── monitors/
│   ├── whale_tracker.go     → Threshold-based transfer detection
│   ├── smart_money.go       → Win-rate & profit factor analysis
│   ├── protocol_health.go   → TVL delta & utilization monitoring
│   ├── mempool_monitor.go   → Sandwich/frontrun detection
│   └── token_scanner.go     → Honeypot & rug pull detection
├── api/
│   ├── handlers.go          → Gin request handlers with JSON responses
│   └── middleware.go         → Token bucket rate limiter, API key auth, logging
├── models/
│   └── types.go             → Data structures for all monitor outputs
└── config/
    └── config.go            → Env-based configuration management
```

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or later

### Build & Run

```bash
# Clone the repository
git clone https://github.com/PasukanT/chainwatcher.git
cd chainwatcher

# Download dependencies
go mod tidy

# Build the binary
go build -o chainwatcher .

# Run the HTTP server (default port 8080)
./chainwatcher

# Or run via go
go run main.go cli.go
```

### CLI Mode

```bash
# Run all monitors
CHAINWATCHER_CLI=1 ./chainwatcher --all --json

# Run specific monitors
CHAINWATCHER_CLI=1 ./chainwatcher --whale --threshold 50000
CHAINWATCHER_CLI=1 ./chainwatcher --smart-money --window 7200
CHAINWATCHER_CLI=1 ./chainwatcher --protocol --json
CHAINWATCHER_CLI=1 ./chainwatcher --mempool --window 300
CHAINWATCHER_CLI=1 ./chainwatcher --tokens --json
```

### API Examples

```bash
# Whale alerts (requires API key)
curl -H "X-API-Key: demo-key-001" http://localhost:8080/api/whales?threshold=100000&window=3600

# Smart money signals
curl -H "X-API-Key: demo-key-001" http://localhost:8080/api/smart-money?window=3600

# Protocol health metrics
curl -H "X-API-Key: demo-key-001" http://localhost:8080/api/health

# Mempool analysis
curl -H "X-API-Key: demo-key-001" http://localhost:8080/api/mempool?window=300

# Token risk scanner
curl -H "X-API-Key: demo-key-001" http://localhost:8080/api/tokens?window=3600

# Public health check (no auth needed)
curl http://localhost:8080/ping
```

---

## ⚙️ Configuration

All configuration via environment variables:

| Variable | Default | Description |
|---|---|---|
| `CW_PORT` | `8080` | Server port |
| `CW_RPC_URL` | Alchemy demo | Ethereum RPC endpoint |
| `CW_API_KEYS` | `demo-key-001` | Comma-separated valid API keys |
| `CW_RATE_LIMIT_RPS` | `10` | Requests per second per IP |
| `CW_RATE_LIMIT_BURST` | `20` | Burst capacity per IP |
| `CW_WHALE_THRESHOLD` | `100000` | Whale alert threshold (USD) |
| `CW_ACCUM_WINDOW` | `3600` | Accumulation detection window (seconds) |
| `CW_TVL_CHANGE_PCT` | `10` | TVL change alert threshold (%) |
| `CW_BAD_DEBT_RATIO` | `0.05` | Bad debt critical ratio |
| `CW_GAS_MULTIPLIER` | `1.5` | Gas anomaly multiplier |
| `CW_HONEYPOT_TAX` | `10` | Tax rate flagged as honeypot (%) |

---

## 📄 License

MIT License — see [LICENSE](LICENSE)
