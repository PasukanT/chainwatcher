package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"chainwatcher/internal/config"
	"chainwatcher/internal/monitors"
)

func runCLI(args []string) {
	fs := flag.NewFlagSet("chainwatcher", flag.ExitOnError)

	whale := fs.Bool("whale", false, "Run whale tracker scan")
	smart := fs.Bool("smart-money", false, "Run smart money analysis")
	protocol := fs.Bool("protocol", false, "Run protocol health check")
	mempool := fs.Bool("mempool", false, "Run mempool monitor scan")
	token := fs.Bool("tokens", false, "Run token scanner")
	all := fs.Bool("all", false, "Run all monitors")
	threshold := fs.Float64("threshold", 100000, "Whale alert threshold in USD")
	duration := fs.Int("window", 3600, "Analysis window in seconds")
	jsonOut := fs.Bool("json", false, "Output as JSON")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	cfg := config.Load()
	_ = cfg // Use config for RPC URL in real implementation

	if *all || *whale {
		result := monitors.ScanWhaleTransfers(*threshold, *duration)
		printResult("Whale Tracker", result, *jsonOut)
	}
	if *all || *smart {
		result := monitors.AnalyzeSmartMoney(*duration)
		printResult("Smart Money", result, *jsonOut)
	}
	if *all || *protocol {
		result := monitors.CheckProtocolHealth()
		printResult("Protocol Health", result, *jsonOut)
	}
	if *all || *mempool {
		result := monitors.AnalyzeMempool(*duration)
		printResult("Mempool Monitor", result, *jsonOut)
	}
	if *all || *token {
		result := monitors.ScanNewTokens(*duration)
		printResult("Token Scanner", result, *jsonOut)
	}
}

func printResult(title string, data interface{}, jsonOut bool) {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(data)
		return
	}
	fmt.Printf("\n═══ %s ═══\n", title)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}
