package soroban

import (
	"context"
	"testing"
)

func TestGetNetwork(t *testing.T) {
	tests := []struct {
		name          string
		result        string
		wantPassport  string
		wantProtocolV uint32
	}{
		{
			name:          "mainnet-like network",
			result:        `{"passport":"Public Global Stellar Network ; September 2019","protocolVersion":18}`,
			wantPassport:  "Public Global Stellar Network ; September 2019",
			wantProtocolV: 18,
		},
		{
			name:          "testnet-like network",
			result:        `{"passport":"Testnet ; September 2019","protocolVersion":18}`,
			wantPassport:  "Testnet ; September 2019",
			wantProtocolV: 18,
		},
		{
			name:          "futurenet-like network",
			result:        `{"passport":"Test SDF Future Network ; September 2019","protocolVersion":19}`,
			wantPassport:  "Test SDF Future Network ; September 2019",
			wantProtocolV: 19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, tt.result, &captured))

			resp, err := client.GetNetwork(context.Background())
			if err != nil {
				t.Fatalf("GetNetwork() returned error: %v", err)
			}

			if captured.Method != "getNetwork" {
				t.Errorf("method = %q, want getNetwork", captured.Method)
			}
			if resp.Passport != tt.wantPassport {
				t.Errorf("Passport = %q, want %q", resp.Passport, tt.wantPassport)
			}
			if resp.ProtocolVersion != tt.wantProtocolV {
				t.Errorf("ProtocolVersion = %d, want %d", resp.ProtocolVersion, tt.wantProtocolV)
			}
		})
	}
}

func TestGetVersionInfo(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		wantCoreVersion  string
		wantStellarCoreV string
	}{
		{
			name:             "typical version info",
			result:           `{"coreVersion":"stellar-horizon-v2.0.0-rc1","stellarCoreVersion":"stellar-core-v20.0.0"}`,
			wantCoreVersion:  "stellar-horizon-v2.0.0-rc1",
			wantStellarCoreV: "stellar-core-v20.0.0",
		},
		{
			name:             "empty version fields",
			result:           `{"coreVersion":"","stellarCoreVersion":""}`,
			wantCoreVersion:  "",
			wantStellarCoreV: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, tt.result, &captured))

			resp, err := client.GetVersionInfo(context.Background())
			if err != nil {
				t.Fatalf("GetVersionInfo() returned error: %v", err)
			}

			if captured.Method != "getVersionInfo" {
				t.Errorf("method = %q, want getVersionInfo", captured.Method)
			}
			if resp.CoreVersion != tt.wantCoreVersion {
				t.Errorf("CoreVersion = %q, want %q", resp.CoreVersion, tt.wantCoreVersion)
			}
			if resp.StellarCoreVersion != tt.wantStellarCoreV {
				t.Errorf("StellarCoreVersion = %q, want %q", resp.StellarCoreVersion, tt.wantStellarCoreV)
			}
		})
	}
}
