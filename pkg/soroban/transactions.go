package soroban

import (
	"context"
	"fmt"
)

// TransactionsPaginationOptions specifies pagination settings for getTransactions.
type TransactionsPaginationOptions struct {
	// Cursor is an opaque string cursor from a previous response to resume pagination.
	Cursor string `json:"cursor,omitempty"`

	// Limit specifies the maximum number of transactions to return (1-200, default 50).
	Limit uint32 `json:"limit,omitempty"`
}

// GetTransactionsRequest holds the arguments for getTransactions.
type GetTransactionsRequest struct {
	// StartLedger is the ledger sequence number to start fetching transactions from (inclusive).
	// Must be omitted if a pagination cursor is provided.
	StartLedger uint32 `json:"startLedger,omitempty"`

	// Pagination specifies optional pagination settings.
	Pagination *TransactionsPaginationOptions `json:"pagination,omitempty"`

	// XDRFormat specifies the response format: "base64" (default) or "json".
	XDRFormat string `json:"xdrFormat,omitempty"`
}

// TransactionInfo represents a single transaction returned by getTransactions.
type TransactionInfo struct {
	// Status is the transaction application status: SUCCESS or FAILED.
	Status string `json:"status"`

	// ApplicationOrder is the index of the transaction within the ledger.
	ApplicationOrder int32 `json:"applicationOrder,omitempty"`

	// FeeBump indicates whether this is a fee-bumped transaction.
	FeeBump bool `json:"feeBump,omitempty"`

	// EnvelopeXDR is the base64 XDR TransactionEnvelope.
	EnvelopeXDR string `json:"envelopeXdr"`

	// ResultXDR is the base64 XDR TransactionResult.
	ResultXDR string `json:"resultXdr"`

	// ResultMetaXDR is the base64 XDR TransactionMeta with ledger entries modified.
	ResultMetaXDR string `json:"resultMetaXdr"`

	// DiagnosticEventsXDR holds diagnostic events emitted during execution.
	DiagnosticEventsXDR []string `json:"diagnosticEventsXdr,omitempty"`

	// Ledger is the ledger sequence number the transaction was applied in.
	Ledger uint32 `json:"ledger"`

	// CreatedAt is the UNIX timestamp when the transaction was applied.
	CreatedAt uint64 `json:"createdAt"`
}

// GetTransactionsResponse is the result of getTransactions.
type GetTransactionsResponse struct {
	// Transactions is the list of transactions retrieved.
	Transactions []TransactionInfo `json:"transactions"`

	// LatestLedger is the sequence number of the latest ledger known to the RPC node.
	LatestLedger uint32 `json:"latestLedger"`

	// LatestLedgerCloseTime is the unix timestamp of the latest ledger close time.
	LatestLedgerCloseTime string `json:"latestLedgerCloseTime,omitempty"`

	// OldestLedger is the sequence number of the oldest ledger in history retention.
	OldestLedger uint32 `json:"oldestLedger,omitempty"`

	// OldestLedgerCloseTime is the unix timestamp of the oldest ledger close time.
	OldestLedgerCloseTime string `json:"oldestLedgerCloseTime,omitempty"`

	// Cursor is an opaque cursor for fetching subsequent pages.
	Cursor string `json:"cursor,omitempty"`
}

// GetTransactions calls getTransactions, returning a list of transactions starting
// from startLedger or continuing from a pagination cursor.
func (c *Client) GetTransactions(ctx context.Context, req GetTransactionsRequest) (*GetTransactionsResponse, error) {
	if req.StartLedger == 0 && (req.Pagination == nil || req.Pagination.Cursor == "") {
		return nil, fmt.Errorf("soroban: getTransactions requires either StartLedger or a Pagination.Cursor")
	}

	var out GetTransactionsResponse
	if err := c.call(ctx, "getTransactions", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
