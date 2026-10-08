package soroban

import (
	"context"
	"fmt"
	"time"
)

// Stellar RPC getTransaction status constants.
const (
	TxStatusSuccess  = "SUCCESS"
	TxStatusNotFound = "NOT_FOUND"
	TxStatusFailed   = "FAILED"
)

// TransactionResponse is the result of getTransaction.
type TransactionResponse struct {
	// Status is the transaction's status: SUCCESS, NOT_FOUND, or FAILED.
	Status string `json:"status"`

	// LatestLedger is the sequence number of the latest ledger known to the RPC node.
	LatestLedger uint32 `json:"latestLedger"`

	// LatestLedgerCloseTime is the unix timestamp of the latest ledger close time.
	LatestLedgerCloseTime string `json:"latestLedgerCloseTime,omitempty"`

	// OldestLedger is the sequence number of the oldest ledger in history retention.
	OldestLedger uint32 `json:"oldestLedger,omitempty"`

	// OldestLedgerCloseTime is the unix timestamp of the oldest ledger close time.
	OldestLedgerCloseTime string `json:"oldestLedgerCloseTime,omitempty"`

	// ApplicationOrder is the index of the transaction within the ledger.
	ApplicationOrder int32 `json:"applicationOrder,omitempty"`

	// FeeBump indicates whether the transaction was fee-bumped.
	FeeBump bool `json:"feeBump,omitempty"`

	// Hash is the transaction hash, hex encoded.
	Hash string `json:"hash,omitempty"`

	// Ledger is the ledger number the transaction was included in.
	Ledger uint32 `json:"ledger,omitempty"`

	// CreatedAt is the UNIX timestamp when the transaction was applied.
	CreatedAt uint64 `json:"createdAt,omitempty"`

	// EnvelopeXDR is the base64 XDR TransactionEnvelope.
	EnvelopeXDR string `json:"envelopeXdr,omitempty"`

	// ResultXDR is the base64 XDR TransactionResult with execution status and return value.
	ResultXDR string `json:"resultXdr,omitempty"`

	// ResultMetaXDR is the base64 XDR TransactionMeta with ledger changes.
	ResultMetaXDR string `json:"resultMetaXdr,omitempty"`

	// DiagnosticEventsXDR holds base64 XDR DiagnosticEvent values recorded during execution.
	DiagnosticEventsXDR []string `json:"diagnosticEventsXdr,omitempty"`

	// FeePaid is the fee actually paid for the transaction, in stroops.
	FeePaid uint32 `json:"feePaid,omitempty"`

	// MaxFee is the maximum fee the transaction was willing to pay, in stroops.
	MaxFee uint32 `json:"maxFee,omitempty"`

	// OperationCount is the number of operations in the transaction.
	OperationCount uint32 `json:"operationCount,omitempty"`

	// FeeMetaXDR is the base64 XDR TransactionMeta for fee calculation.
	FeeMetaXDR string `json:"feeMetaXdr,omitempty"`

	// Memo is the memo associated with the transaction.
	Memo string `json:"memo,omitempty"`

	// Signatures are the transaction signatures.
	Signatures []string `json:"signatures,omitempty"`

	// TimeBounds represents the time bounds of the transaction, if set.
	TimeBounds *TimeBounds `json:"timeBounds,omitempty"`
}

// IsSuccess reports whether the transaction executed successfully.
func (r *TransactionResponse) IsSuccess() bool {
	return r.Status == TxStatusSuccess
}

// IsFailed reports whether the transaction was processed but failed.
func (r *TransactionResponse) IsFailed() bool {
	return r.Status == TxStatusFailed
}

// IsNotFound reports whether the transaction hash was not found.
func (r *TransactionResponse) IsNotFound() bool {
	return r.Status == TxStatusNotFound
}

// TimeBounds represents the optional time bounds of a transaction.
type TimeBounds struct {
	// MinTime is the minimum UNIX timestamp a transaction is valid for.
	MinTime uint64 `json:"minTime"`

	// MaxTime is the maximum UNIX timestamp a transaction is valid for.
	MaxTime uint64 `json:"maxTime"`
}

// GetTransaction calls getTransaction, returning information about a
// previously submitted transaction by its hash.
//
// txHash is the hex-encoded transaction hash (64 characters).
func (c *Client) GetTransaction(ctx context.Context, txHash string) (*TransactionResponse, error) {
	if txHash == "" {
		return nil, fmt.Errorf("soroban: getTransaction requires a transaction hash")
	}

	var out TransactionResponse
	if err := c.call(ctx, "getTransaction", map[string]string{"hash": txHash}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendAndAwaitTransaction submits a signed transaction to the network and polls
// GetTransaction until the transaction is confirmed (status is SUCCESS or FAILED),
// or the context is cancelled.
//
// pollInterval specifies the time to wait between status checks (minimum 500ms).
func (c *Client) SendAndAwaitTransaction(ctx context.Context, envelopeXDR string, pollInterval time.Duration) (*TransactionResponse, error) {
	sendResp, err := c.SendTransaction(ctx, envelopeXDR)
	if err != nil {
		return nil, fmt.Errorf("send transaction: %w", err)
	}
	if sendResp.IsError() {
		return nil, fmt.Errorf("send transaction rejected with error: result_xdr=%s", sendResp.ErrorResultXDR)
	}

	if pollInterval < 500*time.Millisecond {
		pollInterval = 500 * time.Millisecond
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			txResp, err := c.GetTransaction(ctx, sendResp.Hash)
			if err != nil {
				return nil, fmt.Errorf("poll getTransaction: %w", err)
			}
			if txResp.IsSuccess() || txResp.IsFailed() {
				return txResp, nil
			}
		}
	}
}
