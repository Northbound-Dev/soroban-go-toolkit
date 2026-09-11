package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// version is overridden at build time with
// -ldflags "-X main.version=v0.1.0".
var version = "dev"

// options holds the flags every subcommand shares.
type options struct {
	rpcURL  string
	timeout time.Duration
	asJSON  bool
}

// client builds a client from the shared flags. Each command builds its own so
// that a bad --rpc-url is reported when a command actually runs, rather than
// during flag parsing where the error would read as a usage problem.
func (o *options) client() (*soroban.Client, error) {
	return soroban.New(
		soroban.WithURL(o.rpcURL),
		soroban.WithTimeout(o.timeout),
	)
}

func newRootCommand() *cobra.Command {
	opts := &options{}

	root := &cobra.Command{
		Use:     "sorobango",
		Short:   "Query Soroban smart contracts on the Stellar network",
		Version: version,
		Long: `sorobango is a command line client for the Stellar RPC API.

It targets testnet unless --rpc-url says otherwise. There is no free
Stellar-operated RPC endpoint for the public network, so mainnet queries need
the URL of an RPC instance you run yourself or obtain from a provider.`,
		// Errors are printed once by main, and a failed RPC call is not a usage
		// mistake, so cobra should not dump the help text over the result.
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	flags := root.PersistentFlags()
	flags.StringVar(&opts.rpcURL, "rpc-url", soroban.TestnetURL, "Stellar RPC endpoint to query")
	flags.DurationVar(&opts.timeout, "timeout", soroban.DefaultTimeout, "per-request timeout")
	flags.BoolVar(&opts.asJSON, "json", false, "print the raw response as JSON")

	root.AddCommand(
		newHealthCommand(opts),
		newLatestLedgerCommand(opts),
		newLedgerEntriesCommand(opts),
		newContractDataCommand(opts),
		newContractInstanceCommand(opts),
		newSimulateCommand(opts),
		newSendCommand(opts),
		newTransactionCommand(opts),
		newEventsCommand(opts),
	)

	return root
}