package soroban

import "context"

// FeeDistribution contains fee distribution statistics for transactions included in recent ledgers.
type FeeDistribution struct {
	// Max is the maximum inclusion fee observed.
	Max string `json:"max"`

	// Min is the minimum inclusion fee observed.
	Min string `json:"min"`

	// Mode is the most common inclusion fee value.
	Mode string `json:"mode"`

	// P10 through P99 represent fee percentiles.
	P10 string `json:"p10"`
	P20 string `json:"p20"`
	P30 string `json:"p30"`
	P40 string `json:"p40"`
	P50 string `json:"p50"`
	P60 string `json:"p60"`
	P70 string `json:"p70"`
	P80 string `json:"p80"`
	P90 string `json:"p90"`
	P95 string `json:"p95"`
	P99 string `json:"p99"`

	// TransactionCount is the number of transactions included in the distribution.
	TransactionCount uint32 `json:"transactionCount"`

	// LedgerCount is the number of consecutive ledgers used to calculate the distribution.
	LedgerCount uint32 `json:"ledgerCount"`
}

// FeeStatsResponse is the result of getFeeStats.
type FeeStatsResponse struct {
	// SorobanInclusionFee describes statistics for Soroban transactions.
	SorobanInclusionFee FeeDistribution `json:"sorobanInclusionFee"`

	// InclusionFee describes statistics for classic Stellar transactions.
	InclusionFee FeeDistribution `json:"inclusionFee"`

	// LatestLedger is the sequence number of the latest ledger known to the RPC node.
	LatestLedger uint32 `json:"latestLedger"`
}

// GetFeeStats calls getFeeStats, returning fee distribution statistics
// for predicting inclusion fees for future transactions.
func (c *Client) GetFeeStats(ctx context.Context) (*FeeStatsResponse, error) {
	var out FeeStatsResponse
	if err := c.call(ctx, "getFeeStats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
