// Command quickstart demonstrates soroban-go-toolkit against the live Stellar
// testnet.
//
// With no arguments it checks the endpoint's health and reads the latest ledger:
//
//	go run ./examples/quickstart
//
// Pass a contract address to also read that contract's instance entry, which
// confirms whether a contract is deployed there:
//
//	go run ./examples/quickstart -contract CXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
//
// Every call here reaches the real network, so this needs an internet
// connection. It only ever reads: nothing is signed and nothing is submitted.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

func main() {
	rpcURL := flag.String("rpc-url", soroban.TestnetURL, "Stellar RPC endpoint to query")
	contractID := flag.String("contract", "", "optional contract address (C...) whose instance entry to read")
	flag.Parse()

	if err := run(*rpcURL, *contractID); err != nil {
		fmt.Fprintln(os.Stderr, "quickstart:", err)
		os.Exit(1)
	}
}

func run(rpcURL, contractID string) error {
	client, err := soroban.New(soroban.WithURL(rpcURL))
	if err != nil {
		return err
	}

	// One deadline covers the whole run, so a stalled endpoint cannot hang the
	// example indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	fmt.Printf("endpoint: %s\n\n", client.Endpoint())

	if err := showHealth(ctx, client); err != nil {
		return err
	}
	if err := showLatestLedger(ctx, client); err != nil {
		return err
	}

	if contractID == "" {
		fmt.Println("\nPass -contract C... to also read a contract's instance entry.")
		return nil
	}
	return showContractInstance(ctx, client, contractID)
}

func showHealth(ctx context.Context, client *soroban.Client) error {
	health, err := client.GetHealth(ctx)
	if err != nil {
		return fmt.Errorf("getHealth: %w", err)
	}

	fmt.Printf("health:        %s\n", health.Status)
	// Servers on older builds omit the ledger range.
	if health.LatestLedger > 0 {
		fmt.Printf("ledgers held:  %d to %d\n", health.OldestLedger, health.LatestLedger)
	}

	if !health.Healthy() {
		return fmt.Errorf("endpoint reports status %q", health.Status)
	}
	return nil
}

func showLatestLedger(ctx context.Context, client *soroban.Client) error {
	ledger, err := client.GetLatestLedger(ctx)
	if err != nil {
		return fmt.Errorf("getLatestLedger: %w", err)
	}

	fmt.Printf("latest ledger: %d\n", ledger.Sequence)
	fmt.Printf("protocol:      %d\n", ledger.ProtocolVersion)
	fmt.Printf("ledger hash:   %s\n", ledger.ID)
	return nil
}

func showContractInstance(ctx context.Context, client *soroban.Client, contractID string) error {
	fmt.Printf("\nreading the instance entry for %s\n", contractID)

	data, err := client.GetContractInstance(ctx, contractID)
	if errors.Is(err, soroban.ErrEntryNotFound) {
		// Worth distinguishing from a transport failure: the call succeeded and
		// the answer is that nothing is deployed there.
		return fmt.Errorf("no contract is deployed at that address on this network")
	}
	if err != nil {
		return fmt.Errorf("getContractData: %w", err)
	}

	fmt.Printf("  deployed:      yes\n")
	fmt.Printf("  value type:    %s\n", data.Value().Type)
	fmt.Printf("  last modified: ledger %d\n", data.LastModifiedLedgerSeq)
	if data.LiveUntilLedgerSeq != nil {
		fmt.Printf("  live until:    ledger %d\n", *data.LiveUntilLedgerSeq)
	}
	fmt.Printf("  read at:       ledger %d\n", data.LatestLedger)
	return nil
}
