package soroban

import "context"

// NetworkResponse is the result of getNetwork.
type NetworkResponse struct {
	// Passport is the network identifier (e.g., "Testnet ; September 2019").
	Passport string `json:"passport"`

	// ProtocolVersion is the Stellar protocol version the network is using.
	ProtocolVersion uint32 `json:"protocolVersion"`
}

// GetNetwork calls getNetwork, returning information about the connected network.
//
// The passport field identifies the network (e.g., "Testnet ; September 2019").
// The protocol version indicates which Stellar protocol version the network is using.
func (c *Client) GetNetwork(ctx context.Context) (*NetworkResponse, error) {
	var out NetworkResponse
	if err := c.call(ctx, "getNetwork", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VersionInfoResponse is the result of getVersionInfo.
type VersionInfoResponse struct {
	// CoreVersion is the version of the horizon core.
	CoreVersion string `json:"coreVersion"`

	// StellarCoreVersion is the version of stellar-core.
	StellarCoreVersion string `json:"stellarCoreVersion"`
}

// GetVersionInfo calls getVersionInfo, returning version information about the
// connected Stellar core and horizon instances.
func (c *Client) GetVersionInfo(ctx context.Context) (*VersionInfoResponse, error) {
	var out VersionInfoResponse
	if err := c.call(ctx, "getVersionInfo", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}