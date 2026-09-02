package soroban

import "context"

// HealthStatusHealthy is the status string a healthy RPC server reports.
const HealthStatusHealthy = "healthy"

// HealthResponse is the result of getHealth.
type HealthResponse struct {
	// Status is "healthy" on a server that is keeping up with the network.
	Status string `json:"status"`

	// LatestLedger, OldestLedger and LedgerRetentionWindow describe the range
	// of ledgers this server can answer queries about. They were added after
	// the method's first release, so a server running an older build omits
	// them and leaves these zero.
	LatestLedger          uint32 `json:"latestLedger"`
	OldestLedger          uint32 `json:"oldestLedger"`
	LedgerRetentionWindow uint32 `json:"ledgerRetentionWindow"`
}

// Healthy reports whether the server considers itself healthy.
func (h *HealthResponse) Healthy() bool { return h.Status == HealthStatusHealthy }

// GetHealth calls getHealth. It is the cheapest way to check that an endpoint
// is reachable and serving, and takes no parameters.
func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	var out HealthResponse
	if err := c.call(ctx, "getHealth", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
