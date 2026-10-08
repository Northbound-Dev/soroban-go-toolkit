package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

type feeStatsOutput struct {
	LatestLedger        uint32                  `json:"latestLedger"`
	SorobanInclusionFee soroban.FeeDistribution `json:"sorobanInclusionFee"`
	InclusionFee        soroban.FeeDistribution `json:"inclusionFee"`
}

func printFeeStats(w io.Writer, out feeStatsOutput) {
	fmt.Fprintf(w, "latest ledger:               %d\n", out.LatestLedger)
	fmt.Fprintln(w, "\n--- Soroban Transactions ---")
	fmt.Fprintf(w, "transaction count:           %d (across %d ledgers)\n", out.SorobanInclusionFee.TransactionCount, out.SorobanInclusionFee.LedgerCount)
	fmt.Fprintf(w, "min / mode / max:            %s / %s / %s stroops\n", out.SorobanInclusionFee.Min, out.SorobanInclusionFee.Mode, out.SorobanInclusionFee.Max)
	fmt.Fprintf(w, "p50 (median) / p90 / p99:    %s / %s / %s stroops\n", out.SorobanInclusionFee.P50, out.SorobanInclusionFee.P90, out.SorobanInclusionFee.P99)
	fmt.Fprintln(w, "\n--- Classic Transactions ---")
	fmt.Fprintf(w, "transaction count:           %d (across %d ledgers)\n", out.InclusionFee.TransactionCount, out.InclusionFee.LedgerCount)
	fmt.Fprintf(w, "min / mode / max:            %s / %s / %s stroops\n", out.InclusionFee.Min, out.InclusionFee.Mode, out.InclusionFee.Max)
	fmt.Fprintf(w, "p50 (median) / p90 / p99:    %s / %s / %s stroops\n", out.InclusionFee.P50, out.InclusionFee.P90, out.InclusionFee.P99)
}

func newFeeStatsCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "fee-stats",
		Short: "Retrieve transaction fee distribution statistics",
		Long: `Get statistics about inclusion fees from recent ledgers to predict appropriate
transaction fees for Soroban and classic transactions.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			stats, err := client.GetFeeStats(cmd.Context())
			if err != nil {
				return err
			}

			out := feeStatsOutput{
				LatestLedger:        stats.LatestLedger,
				SorobanInclusionFee: stats.SorobanInclusionFee,
				InclusionFee:        stats.InclusionFee,
			}

			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printFeeStats(w, out)
			})
		},
	}
}
