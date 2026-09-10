package soroban

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetTransaction(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		wantHash   string
		wantLedger uint32
		wantFee    uint32
	}{
		{
			name:       "successful transaction",
			result:     `{"latestLedger":1339385,"hash":"c73c5eac58a441d4eb733c35253ae85f783e018f7be5ef974258fed067aabb36","ledger":1339385,"createdAt":1609459200,"feePaid":100,"maxFee":1000,"operationCount":2,"envelopeXDR":"AAAAAgAAAACY7juZTAAAQQRYUDDCwtd3KwAAAADWTNAAAAEAAAAAAAAAAAAAAABAAAAAQAAAAAAAAAB//////////+wAAAAA=","resultMetaXDR":"AAAAAgAAAAMAAAABAQAAAAAAAAAB///////wAAAAA=","feeMetaXDR":"AAAAAgAAAAMAAAABAQAAAAAAAAAB///////wAAAAA=","memo":"","signatures":[],"timeBounds":null}`,
			wantHash:   "c73c5eac58a441d4eb733c35253ae85f783e018f7be5ef974258fed067aabb36",
			wantLedger: 1339385,
			wantFee:    100,
		},
		{
			name:       "transaction with memo and signatures",
			result:     `{"latestLedger":1339386,"hash":"a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890","ledger":1339386,"createdAt":1609459260,"feePaid":150,"maxFee":200,"operationCount":3,"envelopeXDR":"AAAAAgAAAACY7juZTAAAQQRYUDDCwtd3KwAAAADWTNAAAAEAAAAAAAAAAAAAAABAAAAAQAAAAAAAAAB//////////+wAAAAA=","resultMetaXDR":"AAAAAgAAAAMAAAABAQAAAAAAAAAB///////wAAAAA=","feeMetaXDR":"AAAAAgAAAAMAAAABAQAAAAAAAAAB///////wAAAAA=","memo":"Hello Stellar!","signatures":["signature1","signature2"],"timeBounds":{"minTime":1609459200,"maxTime":1609462800}}`,
			wantHash:   "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890",
			wantLedger: 1339386,
			wantFee:    150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, tt.result, &captured))

			resp, err := client.GetTransaction(context.Background(), tt.wantHash)
			if err != nil {
				t.Fatalf("GetTransaction() returned error: %v", err)
			}

			if captured.Method != "getTransaction" {
				t.Errorf("method = %q, want getTransaction", captured.Method)
			}

			var params map[string]string
			if err := json.Unmarshal(captured.Params, &params); err != nil {
				t.Fatalf("could not decode params %s: %v", captured.Params, err)
			}
			if params["hash"] != tt.wantHash {
				t.Errorf("expected hash %q in params, got %q", tt.wantHash, params["hash"])
			}

			if resp.Hash != tt.wantHash {
				t.Errorf("Hash = %q, want %q", resp.Hash, tt.wantHash)
			}
			if resp.LatestLedger != tt.wantLedger {
				t.Errorf("LatestLedger = %d, want %d", resp.LatestLedger, tt.wantLedger)
			}
			if resp.FeePaid != tt.wantFee {
				t.Errorf("FeePaid = %d, want %d", resp.FeePaid, tt.wantFee)
			}
		})
	}
}

func TestGetTransactionRequiresHash(t *testing.T) {
	var captured capturedRequest
	client := newTestClient(t, serveResult(t, `{}`, &captured))

	if _, err := client.GetTransaction(context.Background(), ""); err == nil {
		t.Fatal("GetTransaction() returned nil error, want a validation failure")
	}
	// Error is expected, so we don't check captured request
}
