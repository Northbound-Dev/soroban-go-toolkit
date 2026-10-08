package soroban

import (
	"context"
	"fmt"
)

// Stellar RPC sendTransaction status constants.
const (
	SendStatusPending       = "PENDING"
	SendStatusDuplicate     = "DUPLICATE"
	SendStatusTryAgainLater = "TRY_AGAIN_LATER"
	SendStatusError         = "ERROR"
)

// SendTransactionResponse is the result of sendTransaction.
type SendTransactionResponse struct {
	// Status is the transaction's submission status: PENDING, DUPLICATE,
	// TRY_AGAIN_LATER, or ERROR.
	Status string `json:"status,omitempty"`

	// Hash is the transaction hash, hex encoded.
	Hash string `json:"hash"`

	// LatestLedger is the sequence number of the latest ledger known to the RPC node.
	LatestLedger uint32 `json:"latestLedger"`

	// LatestLedgerCloseTime is the unix timestamp of the latest ledger close time.
	LatestLedgerCloseTime string `json:"latestLedgerCloseTime,omitempty"`

	// ErrorResultXDR is returned if status is ERROR, providing the base64-encoded
	// TransactionResult XDR.
	ErrorResultXDR string `json:"errorResultXdr,omitempty"`

	// DiagnosticEventsXDR holds base64 XDR DiagnosticEvent values explaining why
	// a transaction submission failed.
	DiagnosticEventsXDR []string `json:"diagnosticEventsXdr,omitempty"`

	// Legacy / extended fields:
	FeeCharged     uint32 `json:"feeCharged,omitempty"`
	MemoXDR        string `json:"memoXdr,omitempty"`
	SorobanMetaXDR string `json:"sorobanMetaXdr,omitempty"`
	ResultXDR      string `json:"resultXdr,omitempty"`
	FeeMetaXDR     string `json:"feeMetaXdr,omitempty"`
}

// IsPending reports whether the transaction was accepted into the mempool.
func (r *SendTransactionResponse) IsPending() bool {
	return r.Status == SendStatusPending
}

// IsError reports whether the transaction submission was rejected.
func (r *SendTransactionResponse) IsError() bool {
	return r.Status == SendStatusError
}

// IsDuplicate reports whether the transaction was already submitted.
func (r *SendTransactionResponse) IsDuplicate() bool {
	return r.Status == SendStatusDuplicate
}

// IsTryAgainLater reports whether the node temporarily rejected the transaction.
func (r *SendTransactionResponse) IsTryAgainLater() bool {
	return r.Status == SendStatusTryAgainLater
}

// SendTransaction calls sendTransaction to submit a signed transaction
// to the Stellar network for inclusion in a ledger.
//
// envelopeXDR is a base64 XDR TransactionEnvelope that must be fully signed.
// The transaction will be validated and, if valid, enqueued for ledger inclusion.
//
// To check whether the transaction was successfully processed by the ledger,
// poll GetTransaction or use SendAndAwaitTransaction.
func (c *Client) SendTransaction(ctx context.Context, envelopeXDR string) (*SendTransactionResponse, error) {
	if envelopeXDR == "" {
		return nil, fmt.Errorf("soroban: sendTransaction requires a transaction envelope")
	}

	var out SendTransactionResponse
	// Send "transaction" (standard Stellar RPC) and "tx" (legacy alias)
	params := map[string]string{
		"transaction": envelopeXDR,
		"tx":          envelopeXDR,
	}
	if err := c.call(ctx, "sendTransaction", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
