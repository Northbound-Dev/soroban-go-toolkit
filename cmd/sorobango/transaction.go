package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// transactionOutput is the CLI's own JSON shape for a transaction read.
//
// The client returns XDR types, which would serialise into something verbose
// and tied to the SDK's internal representation. Declaring the output
// explicitly keeps --json a stable contract for scripts.
type transactionOutput struct {
	LatestLedger   uint32 `json:"latestLedger"`
	Hash           string `json:"hash"`
	Ledger         uint32 `json:"ledger"`
	CreatedAt      uint64 `json:"createdAt"`
	FeePaid        uint32 `json:"feePaid"`
	MaxFee         uint32 `json:"maxFee"`
	OperationCount uint32 `json:"operationCount"`
	EnvelopeXDR    string `json:"envelopeXdr"`
	ResultMetaXDR  string `json:"resultMetaXdr"`
	FeeMetaXDR     string `json:"feeMetaXdr"`
	Memo           string `json:"memo"`
	Signatures     []string `json:"signatures,omitempty"`
	TimeBounds     *TimeBoundsOutput `json:"timeBounds,omitempty"`
}

// TimeBoundsOutput is the CLI's own JSON shape for time bounds.
type TimeBoundsOutput struct {
	MinTime uint64 `json:"minTime"`
	MaxTime uint64 `json:"maxTime"`
}

func newTransactionOutput(resp *soroban.TransactionResponse) transactionOutput {
	var timeBounds *TimeBoundsOutput
	if resp.TimeBounds != nil {
		timeBounds = &TimeBoundsOutput{
			MinTime: resp.TimeBounds.MinTime,
			MaxTime: resp.TimeBounds.MaxTime,
		}
	}

	return transactionOutput{
		LatestLedger:   resp.LatestLedger,
		Hash:           resp.Hash,
		Ledger:         resp.Ledger,
		CreatedAt:      resp.CreatedAt,
		FeePaid:        resp.FeePaid,
		MaxFee:         resp.MaxFee,
		OperationCount: resp.OperationCount,
		EnvelopeXDR:    resp.EnvelopeXDR,
		ResultMetaXDR:  resp.ResultMetaXDR,
		FeeMetaXDR:     resp.FeeMetaXDR,
		Memo:           resp.Memo,
		Signatures:     resp.Signatures,
		TimeBounds:     timeBounds,
	}
}

func printTransaction(w io.Writer, out transactionOutput) {
	fmt.Fprintf(w, "latest ledger:   %d\n", out.LatestLedger)
	fmt.Fprintf(w, "hash:            %s\n", out.Hash)
	fmt.Fprintf(w, "ledger:          %d\n", out.Ledger)
	fmt.Fprintf(w, "created at:      %d\n", out.CreatedAt)
	fmt.Fprintf(w, "fee paid:        %d stroops\n", out.FeePaid)
	fmt.Fprintf(w, "max fee:         %d stroops\n", out.MaxFee)
	fmt.Fprintf(w, "operation count: %d\n", out.OperationCount)
	fmt.Fprintf(w, "envelope XDR:    %s\n", out.EnvelopeXDR)
	fmt.Fprintf(w, "result meta XDR: %s\n", out.ResultMetaXDR)
	fmt.Fprintf(w, "fee meta XDR:    %s\n", out.FeeMetaXDR)
	fmt.Fprintf(w, "memo:            %s\n", out.Memo)
	if len(out.Signatures) > 0 {
		fmt.Fprintf(w, "signatures:      %d\n", len(out.Signatures))
		for i, sig := range out.Signatures {
			fmt.Fprintf(w, "  [%d]: %s\n", i, sig)
		}
	}
	if out.TimeBounds != nil {
		fmt.Fprintf(w, "time bounds:\n")
		fmt.Fprintf(w, "  min time: %d\n", out.TimeBounds.MinTime)
		fmt.Fprintf(w, "  max time: %d\n", out.TimeBounds.MaxTime)
	}
}

func newTransactionCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "get-tx <transaction-hash>",
		Short: "Look up a submitted transaction by its hash",
		Long: `Get information about a previously submitted transaction by its hash.

The transaction hash is a 64-character hexadecimal string. This command queries
the RPC node for transaction metadata including fees, timing, and result codes.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txHash := args[0]

			client, err := opts.client()
			if err != nil {
				return err
			}

			resp, err := client.GetTransaction(cmd.Context(), txHash)
			if err != nil {
				return err
			}

			out := newTransactionOutput(resp)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printTransaction(w, out)
			})
		},
	}
}