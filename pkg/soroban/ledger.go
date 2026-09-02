package soroban

import "context"

// LatestLedgerResponse is the result of getLatestLedger.
type LatestLedgerResponse struct {
	// ID is the ledger's hash, hex encoded.
	ID string `json:"id"`

	// ProtocolVersion is the Stellar protocol version the ledger closed under.
	ProtocolVersion uint32 `json:"protocolVersion"`

	// Sequence is the ledger sequence number.
	Sequence uint32 `json:"sequence"`
}

// GetLatestLedger calls getLatestLedger, returning the most recent ledger known
// to the server. It takes no parameters.
func (c *Client) GetLatestLedger(ctx context.Context) (*LatestLedgerResponse, error) {
	var out LatestLedgerResponse
	if err := c.call(ctx, "getLatestLedger", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
