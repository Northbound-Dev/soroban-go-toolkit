package soroban

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
)

// testContractID returns a syntactically valid strkey contract address, built
// rather than hardcoded so the test does not depend on any real contract.
func testContractID(t *testing.T) string {
	t.Helper()

	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i + 1)
	}

	id, err := strkey.Encode(strkey.VersionByteContract, raw)
	if err != nil {
		t.Fatalf("could not build a test contract address: %v", err)
	}
	return id
}

// encodeContractDataEntry builds the base64 XDR LedgerEntryData that a server
// would return for a contract storage entry.
func encodeContractDataEntry(t *testing.T, contractID string, key, val xdr.ScVal) string {
	t.Helper()

	address, err := contractAddress(contractID)
	if err != nil {
		t.Fatalf("contractAddress(%q) returned error: %v", contractID, err)
	}

	data := xdr.LedgerEntryData{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.ContractDataEntry{
			Contract:   address,
			Key:        key,
			Durability: DurabilityPersistent,
			Val:        val,
		},
	}

	encoded, err := xdr.MarshalBase64(data)
	if err != nil {
		t.Fatalf("could not encode contract data entry: %v", err)
	}
	return encoded
}

func TestSymbolKey(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{name: "typical", symbol: "COUNTER"},
		{name: "at the length limit", symbol: "abcdefghijklmnopqrstuvwxyz012345"},
		{name: "empty", symbol: "", wantErr: true},
		{name: "over the length limit", symbol: "abcdefghijklmnopqrstuvwxyz0123456", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := SymbolKey(tt.symbol)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("SymbolKey(%q) returned nil error, want a validation failure", tt.symbol)
				}
				return
			}
			if err != nil {
				t.Fatalf("SymbolKey(%q) returned error: %v", tt.symbol, err)
			}

			if val.Type != xdr.ScValTypeScvSymbol {
				t.Errorf("Type = %s, want ScvSymbol", val.Type)
			}
			if val.Sym == nil {
				t.Fatal("Sym is nil")
			}
			if string(*val.Sym) != tt.symbol {
				t.Errorf("Sym = %q, want %q", string(*val.Sym), tt.symbol)
			}
		})
	}
}

func TestContractDataKeyRoundTrip(t *testing.T) {
	contractID := testContractID(t)

	key, err := SymbolKey("COUNTER")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}

	encoded, err := ContractDataKey(contractID, key, DurabilityPersistent)
	if err != nil {
		t.Fatalf("ContractDataKey() returned error: %v", err)
	}

	// Decoding the key back is what proves it is a well-formed LedgerKey the
	// server would accept, rather than merely a non-empty string.
	var decoded xdr.LedgerKey
	if err := xdr.SafeUnmarshalBase64(encoded, &decoded); err != nil {
		t.Fatalf("could not decode the generated ledger key: %v", err)
	}

	if decoded.Type != xdr.LedgerEntryTypeContractData {
		t.Fatalf("key type = %s, want ContractData", decoded.Type)
	}
	if decoded.ContractData.Durability != DurabilityPersistent {
		t.Errorf("durability = %s, want persistent", decoded.ContractData.Durability)
	}
	if decoded.ContractData.Key.Type != xdr.ScValTypeScvSymbol {
		t.Errorf("key val type = %s, want ScvSymbol", decoded.ContractData.Key.Type)
	}

	// The address must survive the strkey round trip unchanged.
	reEncoded, err := ContractDataKey(contractID, key, DurabilityPersistent)
	if err != nil {
		t.Fatalf("ContractDataKey() second call returned error: %v", err)
	}
	if reEncoded != encoded {
		t.Error("ContractDataKey() is not deterministic for the same inputs")
	}
}

func TestContractDataKeyDurabilityIsHonoured(t *testing.T) {
	contractID := testContractID(t)
	key, err := SymbolKey("COUNTER")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}

	persistent, err := ContractDataKey(contractID, key, DurabilityPersistent)
	if err != nil {
		t.Fatalf("ContractDataKey(persistent) returned error: %v", err)
	}
	temporary, err := ContractDataKey(contractID, key, DurabilityTemporary)
	if err != nil {
		t.Fatalf("ContractDataKey(temporary) returned error: %v", err)
	}

	// Persistent and temporary storage are separate areas; a key that ignored
	// durability would silently read the wrong one.
	if persistent == temporary {
		t.Error("persistent and temporary keys are identical")
	}
}

func TestContractDataKeyRejectsBadAddress(t *testing.T) {
	key, err := SymbolKey("COUNTER")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}

	tests := []struct {
		name       string
		contractID string
	}{
		{name: "empty", contractID: ""},
		{name: "not strkey", contractID: "definitely-not-a-contract"},
		{
			name:       "account address rather than contract",
			contractID: "GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ContractDataKey(tt.contractID, key, DurabilityPersistent); err == nil {
				t.Fatalf("ContractDataKey(%q) returned nil error, want a validation failure", tt.contractID)
			}
		})
	}
}

func TestContractInstanceKey(t *testing.T) {
	contractID := testContractID(t)

	encoded, err := ContractInstanceKey(contractID)
	if err != nil {
		t.Fatalf("ContractInstanceKey() returned error: %v", err)
	}

	var decoded xdr.LedgerKey
	if err := xdr.SafeUnmarshalBase64(encoded, &decoded); err != nil {
		t.Fatalf("could not decode the generated instance key: %v", err)
	}
	if decoded.ContractData.Key.Type != xdr.ScValTypeScvLedgerKeyContractInstance {
		t.Errorf("key val type = %s, want ScvLedgerKeyContractInstance", decoded.ContractData.Key.Type)
	}
}

func TestDecodeContractData(t *testing.T) {
	contractID := testContractID(t)

	key, err := SymbolKey("COUNTER")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}
	value := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: func() *xdr.Uint32 { v := xdr.Uint32(42); return &v }()}

	encoded := encodeContractDataEntry(t, contractID, key, value)

	entry, err := DecodeContractData(encoded)
	if err != nil {
		t.Fatalf("DecodeContractData() returned error: %v", err)
	}
	if entry.Val.Type != xdr.ScValTypeScvU32 {
		t.Fatalf("value type = %s, want ScvU32", entry.Val.Type)
	}
	if entry.Val.U32 == nil || *entry.Val.U32 != 42 {
		t.Errorf("value = %v, want 42", entry.Val.U32)
	}
}

func TestDecodeContractDataRejectsWrongEntryType(t *testing.T) {
	// A TTL entry is valid XDR but not contract data, so it must be reported
	// rather than silently yielding a zero entry.
	data := xdr.LedgerEntryData{
		Type: xdr.LedgerEntryTypeTtl,
		Ttl:  &xdr.TtlEntry{LiveUntilLedgerSeq: 1234},
	}
	encoded, err := xdr.MarshalBase64(data)
	if err != nil {
		t.Fatalf("could not encode a TTL entry: %v", err)
	}

	if _, err := DecodeContractData(encoded); err == nil {
		t.Fatal("DecodeContractData() accepted a TTL entry as contract data")
	}
}

func TestDecodeContractDataRejectsGarbage(t *testing.T) {
	if _, err := DecodeContractData("not base64 at all"); err == nil {
		t.Fatal("DecodeContractData() accepted a non-base64 string")
	}
}

func TestGetContractData(t *testing.T) {
	contractID := testContractID(t)

	key, err := SymbolKey("COUNTER")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}
	value := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: func() *xdr.Uint32 { v := xdr.Uint32(7); return &v }()}

	wantKey, err := ContractDataKey(contractID, key, DurabilityPersistent)
	if err != nil {
		t.Fatalf("ContractDataKey() returned error: %v", err)
	}
	entryXDR := encodeContractDataEntry(t, contractID, key, value)

	result := fmt.Sprintf(
		`{"entries":[{"key":%q,"xdr":%q,"lastModifiedLedgerSeq":1339000,"liveUntilLedgerSeq":1356280}],"latestLedger":1339385}`,
		wantKey, entryXDR,
	)

	var captured capturedRequest
	client := newTestClient(t, serveResult(t, result, &captured))

	data, err := client.GetContractData(context.Background(), contractID, key, DurabilityPersistent)
	if err != nil {
		t.Fatalf("GetContractData() returned error: %v", err)
	}

	// The client must have asked for the key it built, not something else.
	var params struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(captured.Params, &params); err != nil {
		t.Fatalf("could not decode params %s: %v", captured.Params, err)
	}
	if len(params.Keys) != 1 || params.Keys[0] != wantKey {
		t.Errorf("requested keys = %v, want [%s]", params.Keys, wantKey)
	}

	if data.Value().Type != xdr.ScValTypeScvU32 {
		t.Fatalf("value type = %s, want ScvU32", data.Value().Type)
	}
	if data.Value().U32 == nil || *data.Value().U32 != 7 {
		t.Errorf("value = %v, want 7", data.Value().U32)
	}
	if data.LastModifiedLedgerSeq != 1339000 {
		t.Errorf("LastModifiedLedgerSeq = %d, want 1339000", data.LastModifiedLedgerSeq)
	}
	if data.LiveUntilLedgerSeq == nil || *data.LiveUntilLedgerSeq != 1356280 {
		t.Errorf("LiveUntilLedgerSeq = %v, want 1356280", data.LiveUntilLedgerSeq)
	}
	if data.LatestLedger != 1339385 {
		t.Errorf("LatestLedger = %d, want 1339385", data.LatestLedger)
	}
}

func TestGetContractDataNotFound(t *testing.T) {
	contractID := testContractID(t)

	key, err := SymbolKey("MISSING")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}

	client := newTestClient(t, serveResult(t, `{"entries":[],"latestLedger":1339385}`, nil))

	_, err = client.GetContractData(context.Background(), contractID, key, DurabilityPersistent)
	if !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("GetContractData() error = %v, want ErrEntryNotFound", err)
	}
}

func TestGetContractDataRejectsBadAddressBeforeCalling(t *testing.T) {
	key, err := SymbolKey("COUNTER")
	if err != nil {
		t.Fatalf("SymbolKey() returned error: %v", err)
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("client sent a request for an invalid contract address")
	})

	if _, err := client.GetContractData(context.Background(), "nope", key, DurabilityPersistent); err == nil {
		t.Fatal("GetContractData() returned nil error for an invalid contract address")
	}
}

func TestGetContractInstance(t *testing.T) {
	contractID := testContractID(t)

	wantKey, err := ContractInstanceKey(contractID)
	if err != nil {
		t.Fatalf("ContractInstanceKey() returned error: %v", err)
	}

	// A real instance holds an ScvContractInstance; a scalar stands in here
	// because this test is about which key gets requested, not the value shape.
	value := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: func() *xdr.Uint32 { v := xdr.Uint32(1); return &v }()}
	entryXDR := encodeContractDataEntry(t, contractID, contractInstanceKeyVal(), value)

	result := fmt.Sprintf(
		`{"entries":[{"key":%q,"xdr":%q,"lastModifiedLedgerSeq":1339000,"liveUntilLedgerSeq":1356280}],"latestLedger":1339385}`,
		wantKey, entryXDR,
	)

	var captured capturedRequest
	client := newTestClient(t, serveResult(t, result, &captured))

	data, err := client.GetContractInstance(context.Background(), contractID)
	if err != nil {
		t.Fatalf("GetContractInstance() returned error: %v", err)
	}

	// The two instance paths must agree on the key, or reading an instance via
	// GetLedgerEntries and via GetContractInstance would hit different entries.
	var params struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(captured.Params, &params); err != nil {
		t.Fatalf("could not decode params %s: %v", captured.Params, err)
	}
	if len(params.Keys) != 1 || params.Keys[0] != wantKey {
		t.Errorf("requested keys = %v, want [%s]", params.Keys, wantKey)
	}

	if data.LatestLedger != 1339385 {
		t.Errorf("LatestLedger = %d, want 1339385", data.LatestLedger)
	}
}

func TestGetContractInstanceNotDeployed(t *testing.T) {
	contractID := testContractID(t)

	client := newTestClient(t, serveResult(t, `{"entries":[],"latestLedger":1339385}`, nil))

	_, err := client.GetContractInstance(context.Background(), contractID)
	if !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("GetContractInstance() error = %v, want ErrEntryNotFound", err)
	}
}
