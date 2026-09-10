package soroban

import (
	"context"
	"testing"
)

func TestGetEvents(t *testing.T) {
	tests := []struct {
		name       string
		req        *GetEventsRequest
		result     string
		wantLedger uint32
		wantCount  int
	}{
		{
			name:    "no filters",
			req:     &GetEventsRequest{},
			result:  `{"latestLedger":1339385,"events":[{"contractAddress":"CAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX","topics":["abc123"],"data":"AAAAAQ==","ledger":1339385,"ledgerCloseTime":1609459200,"id":"event1","pagingToken":"cursor123"}],"cursor":"cursor123"}`,
			wantLedger: 1339385,
			wantCount:  1,
		},
		{
			name: "with contract filter",
			req: &GetEventsRequest{
				Filters: &EventFilters{
					ContractIDs: []string{"CAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
				},
			},
			result:  `{"latestLedger":1339385,"events":[{"contractAddress":"CAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX","topics":["def456"],"data":"AAAAAg==","ledger":1339385,"ledgerCloseTime":1609459260,"id":"event2","pagingToken":"cursor456"}],"cursor":"cursor456"}`,
			wantLedger: 1339385,
			wantCount:  1,
		},
		{
			name: "with pagination",
			req: &GetEventsRequest{
				Cursor:  "starting_cursor",
				Limit:   10,
				Order:   "asc",
			},
			result:  `{"latestLedger":1339385,"events":[{"contractAddress":"CAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX","topics":["ghi789"],"data":"AAAAAw==","ledger":1339385,"ledgerCloseTime":1609459320,"id":"event3","pagingToken":"cursor789"}],"cursor":"cursor789"}`,
			wantLedger: 1339385,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, tt.result, &captured))

			resp, err := client.GetEvents(context.Background(), tt.req)
			if err != nil {
				t.Fatalf("GetEvents() returned error: %v", err)
			}

			if captured.Method != "getEvents" {
				t.Errorf("method = %q, want getEvents", captured.Method)
			}

			if resp.LatestLedger != tt.wantLedger {
				t.Errorf("LatestLedger = %d, want %d", resp.LatestLedger, tt.wantLedger)
			}
			if len(resp.Events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(resp.Events), tt.wantCount)
			}
		})
	}
}

func TestGetEventsWithNilRequest(t *testing.T) {
	var captured capturedRequest
	client := newTestClient(t, serveResult(t, `{"latestLedger":1339385,"events":[]}`, &captured))

	resp, err := client.GetEvents(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetEvents() returned error: %v", err)
	}

	if captured.Method != "getEvents" {
		t.Errorf("method = %q, want getEvents", captured.Method)
	}
	// Params may be null when request is nil, which is valid JSON-RPC
	if resp.LatestLedger != 1339385 {
		t.Errorf("LatestLedger = %d, want 1339385", resp.LatestLedger)
	}
	if len(resp.Events) != 0 {
		t.Errorf("got %d events, want 0", len(resp.Events))
	}
}