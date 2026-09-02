package soroban

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestGetHealth(t *testing.T) {
	tests := []struct {
		name        string
		result      string
		wantStatus  string
		wantHealthy bool
		wantLatest  uint32
	}{
		{
			name:        "healthy with ledger range",
			result:      `{"status":"healthy","latestLedger":1339385,"oldestLedger":1322105,"ledgerRetentionWindow":17280}`,
			wantStatus:  "healthy",
			wantHealthy: true,
			wantLatest:  1339385,
		},
		{
			// Servers predating the ledger-range fields send status alone; the
			// numeric fields must simply stay zero rather than fail decoding.
			name:        "healthy without ledger range",
			result:      `{"status":"healthy"}`,
			wantStatus:  "healthy",
			wantHealthy: true,
			wantLatest:  0,
		},
		{
			name:        "unhealthy",
			result:      `{"status":"unhealthy","latestLedger":1,"oldestLedger":1,"ledgerRetentionWindow":17280}`,
			wantStatus:  "unhealthy",
			wantHealthy: false,
			wantLatest:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, tt.result, &captured))

			health, err := client.GetHealth(context.Background())
			if err != nil {
				t.Fatalf("GetHealth() returned error: %v", err)
			}

			if captured.Method != "getHealth" {
				t.Errorf("method = %q, want getHealth", captured.Method)
			}
			if health.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", health.Status, tt.wantStatus)
			}
			if health.Healthy() != tt.wantHealthy {
				t.Errorf("Healthy() = %t, want %t", health.Healthy(), tt.wantHealthy)
			}
			if health.LatestLedger != tt.wantLatest {
				t.Errorf("LatestLedger = %d, want %d", health.LatestLedger, tt.wantLatest)
			}
		})
	}
}

func TestGetLatestLedger(t *testing.T) {
	const result = `{"id":"c73c5eac58a441d4eb733c35253ae85f783e018f7be5ef974258fed067aabb36","protocolVersion":28,"sequence":1339385}`

	var captured capturedRequest
	client := newTestClient(t, serveResult(t, result, &captured))

	ledger, err := client.GetLatestLedger(context.Background())
	if err != nil {
		t.Fatalf("GetLatestLedger() returned error: %v", err)
	}

	if captured.Method != "getLatestLedger" {
		t.Errorf("method = %q, want getLatestLedger", captured.Method)
	}
	if len(captured.Params) != 0 {
		t.Errorf("params = %s, want omitted", captured.Params)
	}
	if want := "c73c5eac58a441d4eb733c35253ae85f783e018f7be5ef974258fed067aabb36"; ledger.ID != want {
		t.Errorf("ID = %q, want %q", ledger.ID, want)
	}
	if ledger.ProtocolVersion != 28 {
		t.Errorf("ProtocolVersion = %d, want 28", ledger.ProtocolVersion)
	}
	if ledger.Sequence != 1339385 {
		t.Errorf("Sequence = %d, want 1339385", ledger.Sequence)
	}
}

func TestGetLedgerEntriesSendsKeys(t *testing.T) {
	var captured capturedRequest
	client := newTestClient(t, serveResult(t, `{"entries":[],"latestLedger":1339385}`, &captured))

	keys := []string{"AAAABgAAAAE=", "AAAABgAAAAI="}
	if _, err := client.GetLedgerEntries(context.Background(), keys...); err != nil {
		t.Fatalf("GetLedgerEntries() returned error: %v", err)
	}

	if captured.Method != "getLedgerEntries" {
		t.Errorf("method = %q, want getLedgerEntries", captured.Method)
	}

	var params struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(captured.Params, &params); err != nil {
		t.Fatalf("could not decode params %s: %v", captured.Params, err)
	}
	if len(params.Keys) != len(keys) {
		t.Fatalf("sent %d keys, want %d", len(params.Keys), len(keys))
	}
	for i, key := range keys {
		if params.Keys[i] != key {
			t.Errorf("key %d = %q, want %q", i, params.Keys[i], key)
		}
	}
}

func TestGetLedgerEntriesDecodesEntries(t *testing.T) {
	const result = `{
		"entries":[
			{"key":"AAAABgAAAAE=","xdr":"AAAABgAAAAA=","lastModifiedLedgerSeq":1339000,"liveUntilLedgerSeq":1356280},
			{"key":"AAAAAAAAAAI=","xdr":"AAAAAAAAAAA=","lastModifiedLedgerSeq":1338000}
		],
		"latestLedger":1339385
	}`

	client := newTestClient(t, serveResult(t, result, nil))

	resp, err := client.GetLedgerEntries(context.Background(), "AAAABgAAAAE=", "AAAAAAAAAAI=")
	if err != nil {
		t.Fatalf("GetLedgerEntries() returned error: %v", err)
	}

	if resp.LatestLedger != 1339385 {
		t.Errorf("LatestLedger = %d, want 1339385", resp.LatestLedger)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("decoded %d entries, want 2", len(resp.Entries))
	}

	contractEntry, ok := resp.Get("AAAABgAAAAE=")
	if !ok {
		t.Fatal("Get() did not find the contract data entry")
	}
	if contractEntry.LastModifiedLedgerSeq != 1339000 {
		t.Errorf("LastModifiedLedgerSeq = %d, want 1339000", contractEntry.LastModifiedLedgerSeq)
	}
	ttl, hasTTL := contractEntry.Expires()
	if !hasTTL {
		t.Error("Expires() reported no TTL on an entry carrying liveUntilLedgerSeq")
	}
	if ttl != 1356280 {
		t.Errorf("Expires() = %d, want 1356280", ttl)
	}

	// An account entry has no TTL, so the pointer must stay nil rather than
	// decode to a misleading zero.
	accountEntry, ok := resp.Get("AAAAAAAAAAI=")
	if !ok {
		t.Fatal("Get() did not find the account entry")
	}
	if _, hasTTL := accountEntry.Expires(); hasTTL {
		t.Error("Expires() reported a TTL on an entry without liveUntilLedgerSeq")
	}

	if _, ok := resp.Get("AAAABgAAAAk="); ok {
		t.Error("Get() found an entry for a key that was not returned")
	}
}

func TestGetLedgerEntriesMissingKeysAreOmitted(t *testing.T) {
	// The server omits keys that do not exist, which must read as an empty
	// result rather than an error.
	client := newTestClient(t, serveResult(t, `{"entries":[],"latestLedger":1339385}`, nil))

	resp, err := client.GetLedgerEntries(context.Background(), "AAAABgAAAAE=")
	if err != nil {
		t.Fatalf("GetLedgerEntries() returned error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Errorf("decoded %d entries, want 0", len(resp.Entries))
	}
}

func TestGetLedgerEntriesValidation(t *testing.T) {
	tests := []struct {
		name string
		keys []string
	}{
		{name: "no keys", keys: nil},
		{name: "empty key", keys: []string{"AAAABgAAAAE=", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				t.Error("client sent a request that should have failed validation")
			})

			if _, err := client.GetLedgerEntries(context.Background(), tt.keys...); err == nil {
				t.Fatal("GetLedgerEntries() returned nil error, want a validation failure")
			}
		})
	}
}

func TestSimulateTransactionSendsParams(t *testing.T) {
	tests := []struct {
		name          string
		opts          []SimulateOption
		wantLeeway    uint64
		wantHasConfig bool
		wantAuthMode  string
	}{
		{
			name: "no options omits optional members",
		},
		{
			name:          "instruction leeway",
			opts:          []SimulateOption{WithInstructionLeeway(3000000)},
			wantLeeway:    3000000,
			wantHasConfig: true,
		},
		{
			name:         "auth mode",
			opts:         []SimulateOption{WithAuthMode(AuthModeRecord)},
			wantAuthMode: "record",
		},
		{
			name:          "both",
			opts:          []SimulateOption{WithInstructionLeeway(100), WithAuthMode(AuthModeEnforce)},
			wantLeeway:    100,
			wantHasConfig: true,
			wantAuthMode:  "enforce",
		},
	}

	const envelope = "AAAAAgAAAAA="

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured capturedRequest
			client := newTestClient(t, serveResult(t, `{"latestLedger":1,"minResourceFee":"100"}`, &captured))

			if _, err := client.SimulateTransaction(context.Background(), envelope, tt.opts...); err != nil {
				t.Fatalf("SimulateTransaction() returned error: %v", err)
			}

			if captured.Method != "simulateTransaction" {
				t.Errorf("method = %q, want simulateTransaction", captured.Method)
			}

			var params struct {
				Transaction    string `json:"transaction"`
				ResourceConfig *struct {
					InstructionLeeway uint64 `json:"instructionLeeway"`
				} `json:"resourceConfig"`
				AuthMode string `json:"authMode"`
			}
			if err := json.Unmarshal(captured.Params, &params); err != nil {
				t.Fatalf("could not decode params %s: %v", captured.Params, err)
			}

			if params.Transaction != envelope {
				t.Errorf("transaction = %q, want %q", params.Transaction, envelope)
			}
			if tt.wantHasConfig {
				if params.ResourceConfig == nil {
					t.Fatal("resourceConfig was omitted, want it present")
				}
				if params.ResourceConfig.InstructionLeeway != tt.wantLeeway {
					t.Errorf("instructionLeeway = %d, want %d", params.ResourceConfig.InstructionLeeway, tt.wantLeeway)
				}
			} else if params.ResourceConfig != nil {
				t.Errorf("resourceConfig = %+v, want omitted", params.ResourceConfig)
			}
			if params.AuthMode != tt.wantAuthMode {
				t.Errorf("authMode = %q, want %q", params.AuthMode, tt.wantAuthMode)
			}
		})
	}
}

func TestSimulateTransactionSuccess(t *testing.T) {
	const result = `{
		"latestLedger":1339385,
		"minResourceFee":"58181",
		"transactionData":"AAAAAAAAAAIAAAAG",
		"events":["AAAAAQAAAAA="],
		"results":[{"xdr":"AAAAAwAAABQ=","auth":["AAAAAQAAAAA="]}],
		"stateChanges":[{"type":"updated","key":"AAAABgAAAAE=","before":"AAAAAA==","after":"AAAAAQ=="}]
	}`

	client := newTestClient(t, serveResult(t, result, nil))

	resp, err := client.SimulateTransaction(context.Background(), "AAAAAgAAAAA=")
	if err != nil {
		t.Fatalf("SimulateTransaction() returned error: %v", err)
	}
	if err := resp.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil on a successful simulation", err)
	}

	fee, err := resp.MinResourceFeeInt64()
	if err != nil {
		t.Fatalf("MinResourceFeeInt64() returned error: %v", err)
	}
	if fee != 58181 {
		t.Errorf("MinResourceFeeInt64() = %d, want 58181", fee)
	}

	if resp.NeedsRestore() {
		t.Error("NeedsRestore() = true, want false when restorePreamble is absent")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("decoded %d results, want 1", len(resp.Results))
	}
	if want := "AAAAAwAAABQ="; resp.Results[0].XDR != want {
		t.Errorf("Results[0].XDR = %q, want %q", resp.Results[0].XDR, want)
	}
	if len(resp.Results[0].Auth) != 1 {
		t.Errorf("decoded %d auth entries, want 1", len(resp.Results[0].Auth))
	}
	if len(resp.StateChanges) != 1 {
		t.Fatalf("decoded %d state changes, want 1", len(resp.StateChanges))
	}
	if resp.StateChanges[0].Type != "updated" {
		t.Errorf("StateChanges[0].Type = %q, want updated", resp.StateChanges[0].Type)
	}
}

func TestSimulateTransactionHostFailure(t *testing.T) {
	// A failed simulation is a successful RPC call carrying an error member, so
	// the method must not return a transport error for it.
	const result = `{
		"latestLedger":1339385,
		"minResourceFee":"0",
		"events":["AAAAAQAAAAA="],
		"error":"HostError: Error(Contract, #3)"
	}`

	client := newTestClient(t, serveResult(t, result, nil))

	resp, err := client.SimulateTransaction(context.Background(), "AAAAAgAAAAA=")
	if err != nil {
		t.Fatalf("SimulateTransaction() returned a transport error for a host failure: %v", err)
	}

	simErr := resp.Err()
	if simErr == nil {
		t.Fatal("Err() = nil, want a SimulationError")
	}

	var typed *SimulationError
	if !errors.As(simErr, &typed) {
		t.Fatalf("error %v (%T) is not a *SimulationError", simErr, simErr)
	}
	if typed.Message != "HostError: Error(Contract, #3)" {
		t.Errorf("Message = %q, want the host error", typed.Message)
	}
	// The diagnostic events are the only explanation of the failure, so they
	// must survive onto the error.
	if len(typed.Events) != 1 {
		t.Errorf("carried %d events, want 1", len(typed.Events))
	}
}

func TestSimulateTransactionRestorePreamble(t *testing.T) {
	const result = `{
		"latestLedger":1339385,
		"minResourceFee":"58181",
		"restorePreamble":{"transactionData":"AAAAAAAAAAI=","minResourceFee":"9012"}
	}`

	client := newTestClient(t, serveResult(t, result, nil))

	resp, err := client.SimulateTransaction(context.Background(), "AAAAAgAAAAA=")
	if err != nil {
		t.Fatalf("SimulateTransaction() returned error: %v", err)
	}

	if !resp.NeedsRestore() {
		t.Fatal("NeedsRestore() = false, want true when restorePreamble is present")
	}
	if want := "9012"; resp.RestorePreamble.MinResourceFee != want {
		t.Errorf("RestorePreamble.MinResourceFee = %q, want %q", resp.RestorePreamble.MinResourceFee, want)
	}
}

func TestSimulateTransactionRequiresEnvelope(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("client sent a request for an empty envelope")
	})

	if _, err := client.SimulateTransaction(context.Background(), ""); err == nil {
		t.Fatal("SimulateTransaction() returned nil error, want a validation failure")
	}
}

func TestMinResourceFeeInt64(t *testing.T) {
	tests := []struct {
		name    string
		fee     string
		want    int64
		wantErr bool
	}{
		{name: "typical", fee: "58181", want: 58181},
		{name: "zero", fee: "0", want: 0},
		{name: "large enough to overflow float64 precision", fee: "9007199254740993", want: 9007199254740993},
		{name: "missing", fee: "", wantErr: true},
		{name: "not a number", fee: "cheap", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &SimulateTransactionResponse{MinResourceFee: tt.fee}

			got, err := resp.MinResourceFeeInt64()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("MinResourceFeeInt64() = %d, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("MinResourceFeeInt64() returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("MinResourceFeeInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}
