// Package soroban is a typed Go client for the Stellar RPC JSON-RPC API (the
// service formerly published as Soroban RPC).
//
// Every method returns a Go struct with concrete field types rather than a
// map[string]any, so that response shapes are checked at compile time.
//
// A zero-configuration client targets testnet:
//
//	client, err := soroban.New()
//	if err != nil {
//		return err
//	}
//	health, err := client.GetHealth(context.Background())
//
// Point it elsewhere with WithURL. There is no free Stellar-operated RPC
// endpoint for the public network, so mainnet callers should pass the URL of an
// RPC instance they run themselves or obtain from a provider.
//
// # Errors
//
// Transport and protocol failures are distinguished by type. A *RPCError means
// the server understood the request and refused it, and carries the JSON-RPC
// code. A *HTTPError means the response never reached the JSON-RPC layer.
// Both work with errors.As:
//
//	var rpcErr *soroban.RPCError
//	if errors.As(err, &rpcErr) && rpcErr.Code == -32601 {
//		// method not supported by this server
//	}
//
// # XDR payloads
//
// Several Stellar RPC methods carry XDR-encoded values, which the API transmits
// as base64 strings. Those arrive on two levels.
//
// The raw level exposes them as strings, named so that the XDR type they decode
// to is unambiguous: LedgerEntry.XDR holds a LedgerEntryData,
// SimulateHostFunctionResult.XDR holds an ScVal. Callers already handling XDR
// elsewhere can decode them however they like.
//
// The typed level builds the keys and decodes the values for you, using the
// XDR definitions from github.com/stellar/go. Reading a value from a contract's
// storage needs no XDR knowledge at all:
//
//	key, err := soroban.SymbolKey("COUNTER")
//	if err != nil {
//		return err
//	}
//	data, err := client.GetContractData(ctx, contractID, key, soroban.DurabilityPersistent)
//	if errors.Is(err, soroban.ErrEntryNotFound) {
//		// the contract has not written this key yet
//	}
package soroban
