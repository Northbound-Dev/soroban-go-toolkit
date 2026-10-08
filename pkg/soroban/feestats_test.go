package soroban

import (
	"context"
	"testing"
)

func TestGetFeeStats(t *testing.T) {
	const result = `{
		"sorobanInclusionFee": {
			"max": "50000",
			"min": "100",
			"mode": "100",
			"p10": "100",
			"p20": "100",
			"p30": "100",
			"p40": "100",
			"p50": "100",
			"p60": "120",
			"p70": "150",
			"p80": "200",
			"p90": "500",
			"p95": "1000",
			"p99": "5000",
			"transactionCount": 42,
			"ledgerCount": 10
		},
		"inclusionFee": {
			"max": "10000",
			"min": "100",
			"mode": "100",
			"p10": "100",
			"p20": "100",
			"p30": "100",
			"p40": "100",
			"p50": "100",
			"p60": "100",
			"p70": "100",
			"p80": "100",
			"p90": "100",
			"p95": "100",
			"p99": "200",
			"transactionCount": 120,
			"ledgerCount": 10
		},
		"latestLedger": 1339400
	}`

	var captured capturedRequest
	client := newTestClient(t, serveResult(t, result, &captured))

	stats, err := client.GetFeeStats(context.Background())
	if err != nil {
		t.Fatalf("GetFeeStats() returned error: %v", err)
	}

	if captured.Method != "getFeeStats" {
		t.Errorf("method = %q, want getFeeStats", captured.Method)
	}

	if stats.LatestLedger != 1339400 {
		t.Errorf("LatestLedger = %d, want 1339400", stats.LatestLedger)
	}

	if stats.SorobanInclusionFee.Max != "50000" {
		t.Errorf("Soroban Max = %q, want 50000", stats.SorobanInclusionFee.Max)
	}
	if stats.SorobanInclusionFee.P50 != "100" {
		t.Errorf("Soroban P50 = %q, want 100", stats.SorobanInclusionFee.P50)
	}
	if stats.SorobanInclusionFee.TransactionCount != 42 {
		t.Errorf("Soroban TransactionCount = %d, want 42", stats.SorobanInclusionFee.TransactionCount)
	}
	if stats.InclusionFee.TransactionCount != 120 {
		t.Errorf("InclusionFee TransactionCount = %d, want 120", stats.InclusionFee.TransactionCount)
	}
}
