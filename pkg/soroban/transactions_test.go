package soroban

import (
	"context"
	"net/http"
	"testing"
)

func TestGetTransactions(t *testing.T) {
	const result = `{
		"transactions": [
			{
				"status": "SUCCESS",
				"applicationOrder": 1,
				"feeBump": false,
				"envelopeXdr": "AAAAAgAAAA==",
				"resultXdr": "AAAAAQ==",
				"resultMetaXdr": "AAAAAg==",
				"ledger": 1000,
				"createdAt": 1609459200
			}
		],
		"latestLedger": 1050,
		"latestLedgerCloseTime": "1609460000",
		"oldestLedger": 900,
		"oldestLedgerCloseTime": "1609450000",
		"cursor": "cursor_next_page"
	}`

	var captured capturedRequest
	client := newTestClient(t, serveResult(t, result, &captured))

	req := GetTransactionsRequest{
		StartLedger: 1000,
		Pagination: &TransactionsPaginationOptions{
			Limit: 10,
		},
	}

	resp, err := client.GetTransactions(context.Background(), req)
	if err != nil {
		t.Fatalf("GetTransactions() returned error: %v", err)
	}

	if captured.Method != "getTransactions" {
		t.Errorf("method = %q, want getTransactions", captured.Method)
	}

	if resp.LatestLedger != 1050 {
		t.Errorf("LatestLedger = %d, want 1050", resp.LatestLedger)
	}
	if len(resp.Transactions) != 1 {
		t.Fatalf("len(Transactions) = %d, want 1", len(resp.Transactions))
	}
	tx := resp.Transactions[0]
	if tx.Status != "SUCCESS" {
		t.Errorf("tx.Status = %q, want SUCCESS", tx.Status)
	}
	if tx.Ledger != 1000 {
		t.Errorf("tx.Ledger = %d, want 1000", tx.Ledger)
	}
	if resp.Cursor != "cursor_next_page" {
		t.Errorf("resp.Cursor = %q, want cursor_next_page", resp.Cursor)
	}
}

func TestGetTransactionsValidation(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	// Neither StartLedger nor Cursor provided
	_, err := client.GetTransactions(context.Background(), GetTransactionsRequest{})
	if err == nil {
		t.Fatal("expected error when neither StartLedger nor Cursor is provided")
	}
}
