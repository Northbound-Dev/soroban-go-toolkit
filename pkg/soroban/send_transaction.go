package soroban

import (
	"context"
	"fmt"
)

// SendTransactionResponse is the result of sendTransaction.
type SendTransactionResponse struct {
	// Hash is the transaction hash, hex encoded.
	Hash string `json:"hash"`

	// LatestLedger is the ledger the transaction was included in.
	LatestLedger uint32 `json:"latestLedger"`

	// FeeCharged is the fee actually charged for the transaction, in stroops.
	FeeCharged uint32 `json:"feeCharged"`

	// Memo is the memo is returned if the transaction had a memo.
	MemoXDR string `json:"memoXdr,omitempty"`

	// SorobanMeta is returned if the transaction invoked any smart contracts.
	SorobanMetaXDR string `json:"sorobanMetaXdr,omitempty"`

	// ResultXDR is the base64-encoded TransactionResult.
	ResultXDR string `json:"resultXdr"`

	// FeeMetaXDR is the base64-encoded TransactionMeta for the fee charged.
	FeeMetaXDR string `json:"feeMetaXdr"`
}

// SendTransaction calls sendTransaction to submit a signed transaction
// to the Stellar network for inclusion in a ledger.
//
// envelopeXDR is a base64 XDR TransactionEnvelope that must be fully signed.
// The transaction will be validated and, if valid, submitted to the network.
//
// A successful submission does not guarantee the transaction succeeded -
// it only means the transaction was accepted for inclusion in a ledger.
// To determine if the transaction succeeded, check the ResultXDR field.
func (c *Client) SendTransaction(ctx context.Context, envelopeXDR string) (*SendTransactionResponse, error) {
	if envelopeXDR == "" {
		return nil, fmt.Errorf("soroban: sendTransaction requires a transaction envelope")
	}

	var out SendTransactionResponse
	if err := c.call(ctx, "sendTransaction", map[string]string{"tx": envelopeXDR}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
