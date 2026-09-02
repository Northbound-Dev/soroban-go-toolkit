package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newHealthCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check that the RPC endpoint is reachable and serving",
		Long: `Check that the RPC endpoint is reachable and serving.

Exits non-zero when the endpoint reports anything other than healthy, so this
works as a readiness check in a script or container probe.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			health, err := client.GetHealth(cmd.Context())
			if err != nil {
				return err
			}

			err = emit(cmd, opts.asJSON, health, func(w io.Writer) {
				fmt.Fprintf(w, "status:                  %s\n", health.Status)
				// Older servers omit the ledger range entirely; printing zeroes
				// would imply the endpoint knows about ledger 0.
				if health.LatestLedger > 0 {
					fmt.Fprintf(w, "latest ledger:           %d\n", health.LatestLedger)
					fmt.Fprintf(w, "oldest ledger:           %d\n", health.OldestLedger)
					fmt.Fprintf(w, "ledger retention window: %d\n", health.LedgerRetentionWindow)
				}
			})
			if err != nil {
				return err
			}

			if !health.Healthy() {
				return fmt.Errorf("endpoint reports status %q", health.Status)
			}
			return nil
		},
	}
}

func newLatestLedgerCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "latest-ledger",
		Short: "Show the most recent ledger known to the endpoint",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			ledger, err := client.GetLatestLedger(cmd.Context())
			if err != nil {
				return err
			}

			return emit(cmd, opts.asJSON, ledger, func(w io.Writer) {
				fmt.Fprintf(w, "sequence:         %d\n", ledger.Sequence)
				fmt.Fprintf(w, "hash:             %s\n", ledger.ID)
				fmt.Fprintf(w, "protocol version: %d\n", ledger.ProtocolVersion)
			})
		},
	}
}

func newLedgerEntriesCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "ledger-entries <base64-ledger-key>...",
		Short: "Read raw ledger entries by base64 XDR LedgerKey",
		Long: `Read one or more ledger entries directly.

Keys are base64 XDR LedgerKey values. To read contract storage without building
keys by hand, use the contract-data and contract-instance commands instead.

Keys that do not exist are omitted from the response rather than reported as an
error, so fewer entries than keys means those entries are absent.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			resp, err := client.GetLedgerEntries(cmd.Context(), args...)
			if err != nil {
				return err
			}

			return emit(cmd, opts.asJSON, resp, func(w io.Writer) {
				fmt.Fprintf(w, "latest ledger: %d\n", resp.LatestLedger)
				fmt.Fprintf(w, "found:         %d of %d keys\n", len(resp.Entries), len(args))

				for _, entry := range resp.Entries {
					fmt.Fprintln(w)
					fmt.Fprintf(w, "key:           %s\n", entry.Key)
					fmt.Fprintf(w, "last modified: ledger %d\n", entry.LastModifiedLedgerSeq)
					if ttl, ok := entry.Expires(); ok {
						fmt.Fprintf(w, "live until:    ledger %d\n", ttl)
					}
					fmt.Fprintf(w, "data:          %s\n", entry.XDR)
				}
			})
		},
	}
}
