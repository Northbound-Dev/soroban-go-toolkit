package soroban

import (
	"context"
	"fmt"
)

// LedgerEntry is a single entry returned by getLedgerEntries.
type LedgerEntry struct {
	// Key is the base64 XDR LedgerKey this entry was requested with, echoed
	// back so that callers batching several keys can match results to
	// requests. The server may return entries in any order.
	Key string `json:"key"`

	// XDR is the entry's value: a base64 XDR LedgerEntryData.
	XDR string `json:"xdr"`

	// LastModifiedLedgerSeq is the ledger in which this entry last changed.
	LastModifiedLedgerSeq uint32 `json:"lastModifiedLedgerSeq"`

	// LiveUntilLedgerSeq is the ledger after which the entry expires. It is
	// set only for entries carrying a TTL — contract data and contract code —
	// and nil for entry types that never expire.
	LiveUntilLedgerSeq *uint32 `json:"liveUntilLedgerSeq,omitempty"`
}

// Expires reports whether the entry carries a TTL, and if so the ledger it
// stays live until.
func (e *LedgerEntry) Expires() (uint32, bool) {
	if e.LiveUntilLedgerSeq == nil {
		return 0, false
	}
	return *e.LiveUntilLedgerSeq, true
}

// LedgerEntriesResponse is the result of getLedgerEntries.
type LedgerEntriesResponse struct {
	// Entries holds one element per key that exists in the current ledger.
	// Keys that do not exist are omitted rather than returned empty, so a
	// shorter Entries than the requested key set is normal and means those
	// entries are absent.
	Entries []LedgerEntry `json:"entries"`

	// LatestLedger is the ledger the read was served from.
	LatestLedger uint32 `json:"latestLedger"`
}

// Get returns the entry matching a base64 XDR LedgerKey, and whether it was
// present in the response.
func (r *LedgerEntriesResponse) Get(key string) (*LedgerEntry, bool) {
	for i := range r.Entries {
		if r.Entries[i].Key == key {
			return &r.Entries[i], true
		}
	}
	return nil, false
}

type ledgerEntriesParams struct {
	Keys []string `json:"keys"`
}

// GetLedgerEntries calls getLedgerEntries, the method used to read contract
// state and any other ledger entry directly.
//
// Keys are base64 XDR LedgerKey values. Reading a contract's storage means
// building a LedgerKey of type contractData; see the repository's examples for
// constructing one. Several keys can be read in a single round trip, which is
// the efficient way to load a contract's state.
//
// Keys that do not exist are omitted from the response rather than reported as
// an error, so callers should check the length of Entries or use
// LedgerEntriesResponse.Get.
func (c *Client) GetLedgerEntries(ctx context.Context, keys ...string) (*LedgerEntriesResponse, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("soroban: getLedgerEntries requires at least one key")
	}
	for i, k := range keys {
		if k == "" {
			return nil, fmt.Errorf("soroban: getLedgerEntries key %d is empty", i)
		}
	}

	var out LedgerEntriesResponse
	if err := c.call(ctx, "getLedgerEntries", ledgerEntriesParams{Keys: keys}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
