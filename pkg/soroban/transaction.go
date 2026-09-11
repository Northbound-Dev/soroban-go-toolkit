package soroban

import (
	"context"
	"fmt"
)

// TransactionResponse is the result of getTransaction.
type TransactionResponse struct {
	// LatestLedger is the ledger the transaction was included in.
	LatestLedger uint32 `json:"latestLedger"`

	// Hash is the transaction hash, hex encoded.
	Hash string `json:"hash"`

	// Ledger is the ledger number the transaction was included in.
	Ledger uint32 `json:"ledger"`

	// CreatedAt is the UNIX timestamp when the transaction was created.
	CreatedAt uint64 `json:"createdAt"`

	// FeePaid is the fee actually paid for the transaction, in stroops.
	FeePaid uint32 `json:"feePaid"`

	// MaxFee is the maximum fee the transaction was willing to pay, in stroops.
	MaxFee uint32 `json:"maxFee"`

	// OperationCount is the number of operations in the transaction.
	OperationCount uint32 `json:"operationCount"`

	// EnvelopeXDR is the base64 XDR TransactionEnvelope.
	EnvelopeXDR string `json:"envelopeXdr"`

	// ResultMetaXDR is the base64 XDR TransactionMeta.
	ResultMetaXDR string `json:"resultMetaXdr"`

	// FeeMetaXDR is the base64 XDR TransactionMeta for fee calculation.
	FeeMetaXDR string `json:"feeMetaXdr"`

	// Memo is the memo associated with the transaction.
	Memo string `json:"memo"`

	// Signatures are the transaction signatures.
	Signatures []string `json:"signatures,omitempty"`

	// TimeBounds represents the time bounds of the transaction, if set.
	TimeBounds *TimeBounds `json:"timeBounds,omitempty"`
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
