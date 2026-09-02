package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stellar/go/xdr"
)

// serveRPCResult returns a handler replying with result as the JSON-RPC result
// member, echoing the request id so the client's id check passes.
func serveRPCResult(t *testing.T, result string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID uint64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("server could not decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":%s}`, req.ID, result)
	}
}

// runCommand executes the CLI against an RPC server running handler, returning
// everything it wrote and the error it exited with.
func runCommand(t *testing.T, handler http.HandlerFunc, stdin io.Reader, args ...string) (string, error) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	full := make([]string, 0, len(args)+2)
	full = append(full, args...)
	full = append(full, "--rpc-url", srv.URL)

	var out bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(full)
	if stdin != nil {
		root.SetIn(stdin)
	}

	// Executed before reading the buffer: Go evaluates a return statement's
	// expressions in order, so capturing out.String() alongside the call would
	// read the buffer while it was still empty.
	err := root.ExecuteContext(context.Background())
	return out.String(), err
}

func TestHealthCommand(t *testing.T) {
	const result = `{"status":"healthy","latestLedger":1339385,"oldestLedger":1322105,"ledgerRetentionWindow":17280}`

	out, err := runCommand(t, serveRPCResult(t, result), nil, "health")
	if err != nil {
		t.Fatalf("health returned error: %v\noutput:\n%s", err, out)
	}

	for _, want := range []string{"healthy", "1339385", "17280"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not mention %q:\n%s", want, out)
		}
	}
}

func TestHealthCommandUnhealthyExitsNonZero(t *testing.T) {
	// A reachable but unhealthy endpoint has to fail, or the command is useless
	// as a readiness check.
	const result = `{"status":"unhealthy","latestLedger":1,"oldestLedger":1,"ledgerRetentionWindow":17280}`

	out, err := runCommand(t, serveRPCResult(t, result), nil, "health")
	if err == nil {
		t.Fatalf("health returned nil error for an unhealthy endpoint\noutput:\n%s", out)
	}
	if !strings.Contains(out, "unhealthy") {
		t.Errorf("output does not report the status:\n%s", out)
	}
}

func TestHealthCommandJSON(t *testing.T) {
	const result = `{"status":"healthy","latestLedger":1339385,"oldestLedger":1322105,"ledgerRetentionWindow":17280}`

	out, err := runCommand(t, serveRPCResult(t, result), nil, "health", "--json")
	if err != nil {
		t.Fatalf("health --json returned error: %v", err)
	}

	var decoded struct {
		Status       string `json:"status"`
		LatestLedger uint32 `json:"latestLedger"`
	}
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("--json output is not valid JSON: %v\noutput:\n%s", err, out)
	}
	if decoded.Status != "healthy" {
		t.Errorf("status = %q, want healthy", decoded.Status)
	}
	if decoded.LatestLedger != 1339385 {
		t.Errorf("latestLedger = %d, want 1339385", decoded.LatestLedger)
	}
}

func TestLatestLedgerCommand(t *testing.T) {
	const result = `{"id":"c73c5eac58a441d4eb733c35253ae85f783e018f7be5ef974258fed067aabb36","protocolVersion":28,"sequence":1339385}`

	out, err := runCommand(t, serveRPCResult(t, result), nil, "latest-ledger")
	if err != nil {
		t.Fatalf("latest-ledger returned error: %v", err)
	}

	for _, want := range []string{"1339385", "28", "c73c5eac"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not mention %q:\n%s", want, out)
		}
	}
}

func TestLedgerEntriesCommandReportsMissingKeys(t *testing.T) {
	out, err := runCommand(t, serveRPCResult(t, `{"entries":[],"latestLedger":1339385}`), nil,
		"ledger-entries", "AAAABgAAAAE=")
	if err != nil {
		t.Fatalf("ledger-entries returned error: %v", err)
	}

	// A key that does not exist is not an error, but the output must make the
	// absence obvious rather than printing nothing.
	if !strings.Contains(out, "0 of 1") {
		t.Errorf("output does not report how many keys were found:\n%s", out)
	}
}

func TestSimulateCommandFailureExitsNonZero(t *testing.T) {
	const result = `{"latestLedger":1339385,"minResourceFee":"0","events":["AAAAAQAAAAA="],"error":"HostError: Error(Contract, #3)"}`

	out, err := runCommand(t, serveRPCResult(t, result), nil, "simulate", "AAAAAgAAAAA=")
	if err == nil {
		t.Fatalf("simulate returned nil error for a failed simulation\noutput:\n%s", out)
	}
	if !strings.Contains(out, "simulation failed") {
		t.Errorf("output does not report the failure:\n%s", out)
	}
	if !strings.Contains(out, "diagnostic events") {
		t.Errorf("output omits the diagnostic events, which explain the failure:\n%s", out)
	}
}

func TestSimulateCommandReadsEnvelopeFromStdin(t *testing.T) {
	const envelope = "AAAAAgAAAAABBBBB"

	var got string
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     uint64 `json:"id"`
			Params struct {
				Transaction string `json:"transaction"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("server could not decode request: %v", err)
			return
		}
		got = req.Params.Transaction
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"latestLedger":1,"minResourceFee":"100"}}`, req.ID)
	}

	// Trailing whitespace is what a real "simulate - < file" invocation sends.
	_, err := runCommand(t, handler, strings.NewReader(envelope+"\n"), "simulate", "-")
	if err != nil {
		t.Fatalf("simulate - returned error: %v", err)
	}
	if got != envelope {
		t.Errorf("sent transaction %q, want %q", got, envelope)
	}
}

func TestSimulateCommandRejectsEmptyStdin(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("a request was sent despite an empty envelope on stdin")
	}

	if _, err := runCommand(t, handler, strings.NewReader("\n  \n"), "simulate", "-"); err == nil {
		t.Fatal("simulate - returned nil error for empty stdin")
	}
}

func TestContractDataRejectsUnknownDurability(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("a request was sent despite an invalid --durability")
	}

	_, err := runCommand(t, handler, nil,
		"contract-data", "CA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ", "COUNTER",
		"--durability", "permanent")
	if err == nil {
		t.Fatal("contract-data accepted an unknown durability")
	}
}

func TestParseDurability(t *testing.T) {
	tests := []struct {
		value   string
		want    xdr.ContractDataDurability
		wantErr bool
	}{
		{value: "persistent", want: xdr.ContractDataDurabilityPersistent},
		{value: "temporary", want: xdr.ContractDataDurabilityTemporary},
		{value: "PERSISTENT", want: xdr.ContractDataDurabilityPersistent},
		{value: "permanent", wantErr: true},
		{value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseDurability(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseDurability(%q) = %v, want error", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDurability(%q) returned error: %v", tt.value, err)
			}
			if got != tt.want {
				t.Errorf("parseDurability(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseAuthMode(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{value: "enforce"},
		{value: "record"},
		{value: "record_allow_nonroot"},
		{value: "RECORD"},
		{value: "simulate", wantErr: true},
		{value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseAuthMode(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseAuthMode(%q) = %q, want error", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAuthMode(%q) returned error: %v", tt.value, err)
			}
			if string(got) != strings.ToLower(tt.value) {
				t.Errorf("parseAuthMode(%q) = %q, want the lowercased mode", tt.value, got)
			}
		})
	}
}

func TestFormatScVal(t *testing.T) {
	u32 := xdr.Uint32(42)
	i64 := xdr.Int64(-7)
	boolean := true
	symbol := xdr.ScSymbol("COUNTER")
	text := xdr.ScString("hello")

	tests := []struct {
		name string
		val  xdr.ScVal
		want string
	}{
		{name: "u32", val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u32}, want: "42"},
		{name: "i64", val: xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &i64}, want: "-7"},
		{name: "bool", val: xdr.ScVal{Type: xdr.ScValTypeScvBool, B: &boolean}, want: "true"},
		{name: "symbol", val: xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &symbol}, want: "COUNTER"},
		{name: "string", val: xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &text}, want: "hello"},
		{name: "void", val: xdr.ScVal{Type: xdr.ScValTypeScvVoid}, want: "void"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatScVal(tt.val); got != tt.want {
				t.Errorf("formatScVal() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatScValFallsBackToXDR(t *testing.T) {
	// A structured value has no scalar rendering, so it must come back as
	// base64 XDR rather than an empty string.
	vec := &xdr.ScVec{}
	got := formatScVal(xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &vec})

	if got == "" {
		t.Fatal("formatScVal() returned an empty string for a vector")
	}
	if strings.HasPrefix(got, "<") {
		t.Errorf("formatScVal() could not encode the value: %s", got)
	}
}
