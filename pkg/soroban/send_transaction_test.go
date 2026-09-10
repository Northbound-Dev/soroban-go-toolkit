package soroban

import (
	"context"
	"testing"
)

func TestSendTransaction(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		wantHash   string
		wantLedger uint32
		wantFee    uint32
	}{
		{
			name:    "successful transaction",
			result:  `{"hash":"a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890","latestLedger":1339385,"feeCharged":100,"memoXDR":"AAAAAQ==","sorobanMetaXDR":"AAAAAg==","resultXDR":"AAAAAAAAAQ==","feeMetaXDR":"AAAAAg=="}`,
			wantHash:   "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890",
			wantLedger: 1339385,
			wantFee:    100,
		},
		{
			name:    "transaction with no memo or soroban meta",
			result:  `{"hash":"b2c3d4e5f67890123456789012345678901234567890123456789012345678901c","latestLedger":1339386,"feeCharged":150,"resultXDR":"AAAAAAAAAg==","feeMetaXDR":"AAAAAw=="}`,
			wantHash:   "b2c3d4e5f67890123456789012345678901234567890123456789012345678901c",
			wantLedger: 1339386,
			wantFee:    150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, tt.result, &captured))

			resp, err := client.SendTransaction(context.Background(), "AAAAAgAAAACY7juZTAAAQQRYUDDCwtd3KwAAAADWTNAAAAEAAAAAAAAAAAAAAABAAAAAQAAAAAAAAAB//////////+wAAAAA=")
			if err != nil {
				t.Fatalf("SendTransaction() returned error: %v", err)
			}

			if captured.Method != "sendTransaction" {
				t.Errorf("method = %q, want sendTransaction", captured.Method)
			}

			if resp.Hash != tt.wantHash {
				t.Errorf("Hash = %q, want %q", resp.Hash, tt.wantHash)
			}
			if resp.LatestLedger != tt.wantLedger {
				t.Errorf("LatestLedger = %d, want %d", resp.LatestLedger, tt.wantLedger)
			}
			if resp.FeeCharged != tt.wantFee {
				t.Errorf("FeeCharged = %d, want %d", resp.FeeCharged, tt.wantFee)
			}
		})
	}
}

func TestSendTransactionRequiresEnvelope(t *testing.T) {
	var captured capturedRequest
	client := newTestClient(t, serveResult(t, `{}`, &captured))

	if _, err := client.SendTransaction(context.Background(), ""); err == nil {
		t.Fatal("SendTransaction() returned nil error, want a validation failure")
	}
	// Error is expected from validation, so we don't check captured request
}