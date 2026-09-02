package soroban

import (
	"context"
	"errors"
	"fmt"

	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
)

// ErrEntryNotFound reports a ledger entry that does not exist in the current
// ledger. getLedgerEntries omits missing keys rather than erroring, so this is
// synthesised by the read helpers that expect exactly one entry.
var ErrEntryNotFound = errors.New("soroban: ledger entry not found")

// Durability selects which of a contract's two storage areas to read.
// Persistent entries survive expiry by being restorable; temporary entries do
// not and are gone for good once they expire.
const (
	DurabilityPersistent = xdr.ContractDataDurabilityPersistent
	DurabilityTemporary  = xdr.ContractDataDurabilityTemporary
)

// ContractData is a decoded contract storage entry together with the ledger
// metadata that came back with it.
type ContractData struct {
	// Entry is the decoded contract data, whose Val field holds the stored
	// value.
	Entry xdr.ContractDataEntry

	// LastModifiedLedgerSeq is the ledger in which the entry last changed.
	LastModifiedLedgerSeq uint32

	// LiveUntilLedgerSeq is the ledger the entry stays live until. Contract
	// data always carries a TTL, so this is normally set; it is nil only if a
	// server omits the field.
	LiveUntilLedgerSeq *uint32

	// LatestLedger is the ledger the read was served from.
	LatestLedger uint32
}

// Value returns the stored value.
func (d *ContractData) Value() xdr.ScVal { return d.Entry.Val }

// contractAddress converts a contract's strkey "C..." identifier into the
// ScAddress the ledger key needs.
func contractAddress(contractID string) (xdr.ScAddress, error) {
	raw, err := strkey.Decode(strkey.VersionByteContract, contractID)
	if err != nil {
		return xdr.ScAddress{}, fmt.Errorf("soroban: %q is not a contract address: %w", contractID, err)
	}
	if len(raw) != 32 {
		return xdr.ScAddress{}, fmt.Errorf("soroban: contract address %q decoded to %d bytes, want 32", contractID, len(raw))
	}

	var id xdr.ContractId
	copy(id[:], raw)

	return xdr.ScAddress{
		Type:       xdr.ScAddressTypeScAddressTypeContract,
		ContractId: &id,
	}, nil
}

// ContractDataKey builds the base64 XDR LedgerKey that reads one entry from a
// contract's storage, for passing to GetLedgerEntries.
//
// contractID is a strkey contract address ("C..."). key is the storage key the
// contract writes under — SymbolKey covers the common case of a named value.
func ContractDataKey(contractID string, key xdr.ScVal, durability xdr.ContractDataDurability) (string, error) {
	address, err := contractAddress(contractID)
	if err != nil {
		return "", err
	}

	ledgerKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   address,
			Key:        key,
			Durability: durability,
		},
	}

	encoded, err := xdr.MarshalBase64(ledgerKey)
	if err != nil {
		return "", fmt.Errorf("soroban: encode contract data ledger key: %w", err)
	}
	return encoded, nil
}

// contractInstanceKeyVal is the storage key a contract's instance entry lives
// under. Instances are keyed by a dedicated ScVal type rather than a symbol.
func contractInstanceKeyVal() xdr.ScVal {
	return xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance}
}

// ContractInstanceKey builds the ledger key for a contract's instance entry,
// for callers batching several reads through GetLedgerEntries. To read just the
// instance, GetContractInstance is more direct.
func ContractInstanceKey(contractID string) (string, error) {
	return ContractDataKey(contractID, contractInstanceKeyVal(), DurabilityPersistent)
}

// SymbolKey builds the ScVal for a symbol storage key, the type contracts most
// often use to name a stored value. Symbols are limited to 32 characters.
func SymbolKey(name string) (xdr.ScVal, error) {
	if name == "" {
		return xdr.ScVal{}, fmt.Errorf("soroban: symbol key is empty")
	}
	if len(name) > 32 {
		return xdr.ScVal{}, fmt.Errorf("soroban: symbol key %q is %d characters, the maximum is 32", name, len(name))
	}

	symbol := xdr.ScSymbol(name)
	return xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &symbol}, nil
}

// DecodeContractData decodes the base64 XDR LedgerEntryData carried by a
// LedgerEntry into a contract data entry, failing if the entry is some other
// ledger entry type.
func DecodeContractData(entryXDR string) (xdr.ContractDataEntry, error) {
	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(entryXDR, &data); err != nil {
		return xdr.ContractDataEntry{}, fmt.Errorf("soroban: decode ledger entry data: %w", err)
	}

	contractData, ok := data.GetContractData()
	if !ok {
		return xdr.ContractDataEntry{}, fmt.Errorf("soroban: ledger entry is a %s, not contract data", data.Type)
	}
	return contractData, nil
}

// GetContractData reads a single value from a contract's storage, building the
// ledger key, calling getLedgerEntries, and decoding the result.
//
// It returns ErrEntryNotFound when the entry does not exist, which for
// temporary storage also covers an entry that has expired:
//
//	key, err := soroban.SymbolKey("COUNTER")
//	if err != nil {
//		return err
//	}
//	data, err := client.GetContractData(ctx, contractID, key, soroban.DurabilityPersistent)
//	if errors.Is(err, soroban.ErrEntryNotFound) {
//		// contract has not written COUNTER yet
//	}
func (c *Client) GetContractData(ctx context.Context, contractID string, key xdr.ScVal, durability xdr.ContractDataDurability) (*ContractData, error) {
	ledgerKey, err := ContractDataKey(contractID, key, durability)
	if err != nil {
		return nil, err
	}

	resp, err := c.GetLedgerEntries(ctx, ledgerKey)
	if err != nil {
		return nil, err
	}
	if len(resp.Entries) == 0 {
		return nil, ErrEntryNotFound
	}

	entry := resp.Entries[0]
	contractData, err := DecodeContractData(entry.XDR)
	if err != nil {
		return nil, err
	}

	return &ContractData{
		Entry:                 contractData,
		LastModifiedLedgerSeq: entry.LastModifiedLedgerSeq,
		LiveUntilLedgerSeq:    entry.LiveUntilLedgerSeq,
		LatestLedger:          resp.LatestLedger,
	}, nil
}

// GetContractInstance reads a contract's instance entry, which holds its
// executable reference and instance storage.
//
// This is the cheapest way to confirm a contract is actually deployed at an
// address: ErrEntryNotFound means nothing is there.
func (c *Client) GetContractInstance(ctx context.Context, contractID string) (*ContractData, error) {
	return c.GetContractData(ctx, contractID, contractInstanceKeyVal(), DurabilityPersistent)
}
