package soroban

import (
	"context"
	"fmt"
	"strconv"
)

// AuthMode selects how simulateTransaction treats authorization entries.
type AuthMode string

const (
	// AuthModeEnforce simulates with the authorization entries already present
	// on the transaction, failing if they are insufficient.
	AuthModeEnforce AuthMode = "enforce"

	// AuthModeRecord ignores any supplied entries and records the
	// authorizations the invocation would need, returning them in
	// SimulateHostFunctionResult.Auth.
	AuthModeRecord AuthMode = "record"

	// AuthModeRecordAllowNonRoot records authorizations and additionally
	// permits non-root authorization, which is needed when simulating
	// invocations that authorize on behalf of another address.
	AuthModeRecordAllowNonRoot AuthMode = "record_allow_nonroot"
)

// SimulateHostFunctionResult is the outcome of one host function invocation.
type SimulateHostFunctionResult struct {
	// XDR is the value the host function returned: a base64 XDR ScVal.
	XDR string `json:"xdr"`

	// Auth holds the authorization entries the invocation requires, as base64
	// XDR SorobanAuthorizationEntry values. It is populated when simulating in
	// a recording auth mode, and these entries must be attached to the
	// transaction before submission.
	Auth []string `json:"auth,omitempty"`
}

// RestorePreamble describes a restore operation that must be submitted before
// the simulated transaction can succeed, because the transaction touches
// archived entries. It is non-nil only in that case.
type RestorePreamble struct {
	// TransactionData is a base64 XDR SorobanTransactionData for the restore.
	TransactionData string `json:"transactionData"`

	// MinResourceFee is the restore's resource fee in stroops, encoded as a
	// decimal string.
	MinResourceFee string `json:"minResourceFee"`
}

// LedgerEntryChange records one ledger entry the simulation would alter, which
// is useful for showing a user what a transaction is about to do.
type LedgerEntryChange struct {
	// Type is "created", "updated", or "deleted".
	Type string `json:"type"`

	// Key is the affected entry's base64 XDR LedgerKey.
	Key string `json:"key"`

	// Before and After are base64 XDR LedgerEntry values, empty for a created
	// and a deleted entry respectively.
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// SimulateTransactionResponse is the result of simulateTransaction.
//
// A response with a non-empty Error is a simulation the server ran and which
// failed inside the host; the RPC call itself succeeded. Always check Err
// before trusting the other fields.
type SimulateTransactionResponse struct {
	// LatestLedger is the ledger the simulation ran against.
	LatestLedger uint32 `json:"latestLedger"`

	// MinResourceFee is the resource fee the transaction must pay, in stroops,
	// encoded by the API as a decimal string. Use MinResourceFeeInt64 to parse
	// it.
	MinResourceFee string `json:"minResourceFee"`

	// TransactionData is the base64 XDR SorobanTransactionData holding the
	// footprint and resource values to attach to the real transaction.
	TransactionData string `json:"transactionData,omitempty"`

	// Events holds diagnostic events emitted during simulation, as base64 XDR
	// DiagnosticEvent values.
	Events []string `json:"events,omitempty"`

	// Results holds one entry per host function invoked. A transaction with a
	// single InvokeHostFunction operation yields exactly one.
	Results []SimulateHostFunctionResult `json:"results,omitempty"`

	// RestorePreamble is set when archived entries must be restored first.
	RestorePreamble *RestorePreamble `json:"restorePreamble,omitempty"`

	// StateChanges lists the ledger entries the simulation would alter.
	StateChanges []LedgerEntryChange `json:"stateChanges,omitempty"`

	// Error holds the host's error message when the simulation failed.
	Error string `json:"error,omitempty"`
}

// Err reports a simulation that ran but failed, returning a *SimulationError
// carrying the diagnostic events, or nil when the simulation succeeded.
func (r *SimulateTransactionResponse) Err() error {
	if r.Error == "" {
		return nil
	}
	return &SimulationError{Message: r.Error, Events: r.Events}
}

// MinResourceFeeInt64 parses MinResourceFee, which the API encodes as a decimal
// string so that it survives JSON consumers without 64-bit integers.
func (r *SimulateTransactionResponse) MinResourceFeeInt64() (int64, error) {
	if r.MinResourceFee == "" {
		return 0, fmt.Errorf("soroban: response carried no minResourceFee")
	}
	fee, err := strconv.ParseInt(r.MinResourceFee, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("soroban: parse minResourceFee %q: %w", r.MinResourceFee, err)
	}
	return fee, nil
}

// NeedsRestore reports whether archived entries must be restored before the
// simulated transaction can be submitted.
func (r *SimulateTransactionResponse) NeedsRestore() bool {
	return r.RestorePreamble != nil
}

type resourceConfig struct {
	InstructionLeeway uint64 `json:"instructionLeeway"`
}

type simulateParams struct {
	Transaction    string          `json:"transaction"`
	ResourceConfig *resourceConfig `json:"resourceConfig,omitempty"`
	AuthMode       AuthMode        `json:"authMode,omitempty"`
}

// SimulateOption customises a simulateTransaction call.
type SimulateOption func(*simulateParams)

// WithInstructionLeeway adds headroom to the CPU instruction count the
// simulation reports, which helps a transaction survive small state changes
// between simulating and submitting.
func WithInstructionLeeway(instructions uint64) SimulateOption {
	return func(p *simulateParams) {
		p.ResourceConfig = &resourceConfig{InstructionLeeway: instructions}
	}
}

// WithAuthMode selects the authorization mode. Omitted, the server infers one
// from whether the transaction already carries authorization entries.
func WithAuthMode(mode AuthMode) SimulateOption {
	return func(p *simulateParams) { p.AuthMode = mode }
}

// SimulateTransaction calls simulateTransaction to run a transaction against
// the current ledger without submitting it, returning the resource fee,
// footprint, and any required authorization entries.
//
// envelopeXDR is a base64 XDR TransactionEnvelope. Simulation is a read-only
// operation: nothing is written and the transaction need not be signed.
//
// A failing simulation is reported through the returned response, not as an
// error, because the RPC call itself succeeded. Check Err on the result:
//
//	resp, err := client.SimulateTransaction(ctx, envelope)
//	if err != nil {
//		return err // transport or protocol failure
//	}
//	if err := resp.Err(); err != nil {
//		return err // the contract call itself failed
//	}
func (c *Client) SimulateTransaction(ctx context.Context, envelopeXDR string, opts ...SimulateOption) (*SimulateTransactionResponse, error) {
	if envelopeXDR == "" {
		return nil, fmt.Errorf("soroban: simulateTransaction requires a base64 XDR transaction envelope")
	}

	params := simulateParams{Transaction: envelopeXDR}
	for _, opt := range opts {
		opt(&params)
	}

	var out SimulateTransactionResponse
	if err := c.call(ctx, "simulateTransaction", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
