package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

type transactionsOutput struct {
	Transactions          []soroban.TransactionInfo `json:"transactions"`
	LatestLedger          uint32                    `json:"latestLedger"`
	LatestLedgerCloseTime string                    `json:"latestLedgerCloseTime,omitempty"`
	OldestLedger          uint32                    `json:"oldestLedger,omitempty"`
	OldestLedgerCloseTime string                    `json:"oldestLedgerCloseTime,omitempty"`
	Cursor                string                    `json:"cursor,omitempty"`
}

func printTransactions(w io.Writer, out transactionsOutput) {
	fmt.Fprintf(w, "latest ledger: %d\n", out.LatestLedger)
	if out.OldestLedger > 0 {
		fmt.Fprintf(w, "oldest ledger: %d\n", out.OldestLedger)
	}
	fmt.Fprintf(w, "transactions:  %d returned\n", len(out.Transactions))
	if out.Cursor != "" {
		fmt.Fprintf(w, "next cursor:   %s\n", out.Cursor)
	}
	fmt.Fprintln(w)

	for i, tx := range out.Transactions {
		fmt.Fprintf(w, "[%d] Ledger %d | Status: %s | AppOrder: %d\n", i+1, tx.Ledger, tx.Status, tx.ApplicationOrder)
		fmt.Fprintf(w, "    result XDR:   %s\n", tx.ResultXDR)
		fmt.Fprintf(w, "    envelope XDR: %s\n", tx.EnvelopeXDR)
		if len(tx.DiagnosticEventsXDR) > 0 {
			fmt.Fprintf(w, "    events:       %d diagnostic events\n", len(tx.DiagnosticEventsXDR))
		}
	}
}

func newTransactionsCommand(opts *options) *cobra.Command {
	var startLedger uint32
	var cursor string
	var limit uint32
	var xdrFormat string

	cmd := &cobra.Command{
		Use:   "get-transactions",
		Short: "Query a range of historical transactions from the ledger",
		Long: `Get a list of historical transactions from the RPC history retention window.
Provide --start-ledger to begin querying from a specific ledger sequence, or
--cursor to resume pagination.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if startLedger == 0 && cursor == "" {
				return fmt.Errorf("either --start-ledger or --cursor is required")
			}

			client, err := opts.client()
			if err != nil {
				return err
			}

			req := soroban.GetTransactionsRequest{
				StartLedger: startLedger,
				XDRFormat:   xdrFormat,
			}
			if cursor != "" || limit > 0 {
				req.Pagination = &soroban.TransactionsPaginationOptions{
					Cursor: cursor,
					Limit:  limit,
				}
			}

			resp, err := client.GetTransactions(cmd.Context(), req)
			if err != nil {
				return err
			}

			out := transactionsOutput{
				Transactions:          resp.Transactions,
				LatestLedger:          resp.LatestLedger,
				LatestLedgerCloseTime: resp.LatestLedgerCloseTime,
				OldestLedger:          resp.OldestLedger,
				OldestLedgerCloseTime: resp.OldestLedgerCloseTime,
				Cursor:                resp.Cursor,
			}

			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printTransactions(w, out)
			})
		},
	}

	cmd.Flags().Uint32Var(&startLedger, "start-ledger", 0, "ledger sequence to start fetching from")
	cmd.Flags().StringVar(&cursor, "cursor", "", "pagination cursor to resume from")
	cmd.Flags().Uint32Var(&limit, "limit", 50, "maximum number of transactions to return (1-200)")
	cmd.Flags().StringVar(&xdrFormat, "xdr-format", "base64", "format for XDR fields: base64 or json")

	return cmd
}
