package monitors

import (
	"math"
	"strings"
	"sync"
	"time"

	"chainwatcher/internal/models"
)

// tokenDeployment represents a raw token deployment event.
type tokenDeployment struct {
	Address     string
	Name        string
	Symbol      string
	Deployer    string
	DeployBlock uint64
	Timestamp   time.Time
	Bytecode    []byte
	TxHash      string
}

var (
	tsMu       sync.RWMutex
	tsTokens   []models.TokenInfo
)

// ScanNewTokens detects newly deployed tokens and evaluates them for
// honeypot patterns, rug pull red flags, and suspicious characteristics.
func ScanNewTokens(windowSec int) *models.APIResponse {
	tsMu.RLock()
	defer tsMu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)

	// Fetch recent token deployments from the chain
	deployments := fetchTokenDeployments(cutoff)

	var tokens []models.TokenInfo
	for _, d := range deployments {
		token := analyzeToken(d)
		tokens = append(tokens, token)
	}

	// Sort by risk score descending (highest risk first)
	sortTokensByRisk(tokens)

	return &models.APIResponse{
		Success:   true,
		Data:      tokens,
		Count:     len(tokens),
		Timestamp: now,
	}
}

// analyzeToken evaluates a token deployment for red flags and returns a risk score.
func analyzeToken(d tokenDeployment) models.TokenInfo {
	var flags []string
	riskScore := 0.0

	// Check for honeypot patterns
	if hasNoSellFunction(d.Bytecode) {
		flags = append(flags, "NO_SELL_FUNCTION")
		riskScore += 40
	}

	// Check buy/sell tax (simulated via bytecode analysis)
	buyTax, sellTax := estimateTaxRates(d.Bytecode)
	if sellTax > 10 {
		flags = append(flags, "HIGH_SELL_TAX")
		riskScore += math.Min(sellTax, 30)
	}
	if buyTax > 5 {
		flags = append(flags, "HIGH_BUY_TAX")
		riskScore += math.Min(buyTax/2, 15)
	}

	// Check liquidity lock status
	if !isLiquidityLocked(d.Address) {
		flags = append(flags, "UNLOCKED_LIQUIDITY")
		riskScore += 25
	}

	// Check if ownership is renounced
	isRenounced := isOwnershipRenounced(d.Bytecode)
	if !isRenounced {
		riskScore += 10
	}

	// Check for proxy patterns (upgradeable = potential rug)
	if hasProxyPattern(d.Bytecode) {
		flags = append(flags, "PROXY_CONTRACT")
		riskScore += 15
	}

	// Check deployer history
	if isRepeatDeployer(d.Deployer) {
		flags = append(flags, "REPEAT_DEPLOYER")
		riskScore += 20
	}

	riskScore = math.Min(100, riskScore)

	return models.TokenInfo{
		Address:       d.Address,
		Name:          d.Name,
		Symbol:        d.Symbol,
		Deployer:      d.Deployer,
		DeployBlock:   d.DeployBlock,
		LiquidityUSD:  estimateLiquidity(d.Address),
		BuyTax:        buyTax,
		SellTax:       sellTax,
		HoneypotFlags: flags,
		RiskScore:     math.Round(riskScore*100) / 100,
		IsRenounced:   isRenounced,
		Timestamp:     d.Timestamp,
	}
}

// hasNoSellFunction checks if the contract bytecode lacks a sell/transfer function.
func hasNoSellFunction(bytecode []byte) bool {
	if len(bytecode) == 0 {
		return false
	}
	// Check for ERC20 transfer selector (0xa9059cbb) in bytecode
	code := string(bytecode)
	return !strings.Contains(code, "a9059cbb")
}

// estimateTaxRates simulates buy/sell tax detection from bytecode analysis.
func estimateTaxRates(bytecode []byte) (buyTax, sellTax float64) {
	if len(bytecode) < 10 {
		return 0, 0
	}
	// Simulated: In production, this calls the contract to measure actual transfer amounts
	code := string(bytecode)
	if strings.Contains(code, "tax") {
		return 3.0, 12.0
	}
	return 0, 0
}

// isLiquidityLocked checks if LP tokens are sent to a burn/lock address.
func isLiquidityLocked(address string) bool {
	// In production: check if LP tokens are in a locker contract
	lockedAddresses := map[string]bool{
		"0x000000000000000000000000000000000000dead": true,
		"0x0000000000000000000000000000000000000000": true,
	}
	_ = lockedAddresses
	return len(address)%3 == 0 // Simulated check
}

// isOwnershipRenounced checks if the contract has renounced ownership.
func isOwnershipRenounced(bytecode []byte) bool {
	code := string(bytecode)
	return strings.Contains(code, "renounced")
}

// hasProxyPattern checks for EIP-1967 proxy bytecode patterns.
func hasProxyPattern(bytecode []byte) bool {
	code := string(bytecode)
	return strings.Contains(code, "proxy") || strings.Contains(code, "delegatecall")
}

// isRepeatDeployer checks if this deployer has launched many tokens.
func isRepeatDeployer(deployer string) bool {
	// In production: query deployer's token creation history
	return len(deployer)%2 == 0 // Simulated
}

// estimateLiquidity returns simulated initial liquidity in USD.
func estimateLiquidity(address string) float64 {
	return float64(10000 + len(address)*1000)
}

// sortTokensByRisk orders tokens by risk score descending.
func sortTokensByRisk(tokens []models.TokenInfo) {
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			if tokens[j].RiskScore > tokens[i].RiskScore {
				tokens[i], tokens[j] = tokens[j], tokens[i]
			}
		}
	}
}

// fetchTokenDeployments simulates fetching recent token deploys.
func fetchTokenDeployments(since time.Time) []tokenDeployment {
	now := time.Now()
	return []tokenDeployment{
		{Address: "0xNewToken1a2b3c4d5e6f7890abcdef12345678", Name: "SafeMoonV3", Symbol: "SAFE3", Deployer: "0xScammer1", DeployBlock: 19500000, Timestamp: now.Add(-120 * time.Second), Bytecode: []byte("tax_proxy_delegatecall")},
		{Address: "0xNewToken2a2b3c4d5e6f7890abcdef12345678", Name: "ElonMars", Symbol: "ELON", Deployer: "0xScammer2", DeployBlock: 19500100, Timestamp: now.Add(-60 * time.Second), Bytecode: []byte("no_transfer_renounced")},
		{Address: "0xNewToken3a2b3c4d5e6f7890abcdef12345678", Name: "RealProject", Symbol: "REAL", Deployer: "0xBuilder1", DeployBlock: 19500200, Timestamp: now.Add(-30 * time.Second), Bytecode: []byte("standard_erc20_a9059cbb")},
	}
}
