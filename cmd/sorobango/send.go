package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// readSendEnvelope resolves the envelope argument, treating "-" as standard input.
func readSendEnvelope(cmd *cobra.Command, arg string) (string, error) {
	if arg != "-" {
		return strings.TrimSpace(arg), nil
	}

	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read transaction envelope from stdin: %w", err)
	}

	envelope := strings.TrimSpace(string(raw))
	if envelope == "" {
		return "", fmt.Errorf("no transaction envelope on stdin")
	}
	return envelope, nil
}

// sendOutput is the CLI's own JSON shape for a sent transaction.
//
// The client returns types that would serialise into something verbose
// and tied to the SDK's internal representation. Declaring the output
// explicitly keeps --json a stable contract for scripts.
type sendOutput struct {
	Hash         string `json:"hash"`
	LatestLedger uint32 `json:"latestLedger"`
	FeeCharged   uint32 `json:"feeCharged"`
	MemoXDR      string `json:"memoXdr,omitempty"`
	SorobanMetaXDR string `json:"sorobanMetaXdr,omitempty"`
	ResultXDR    string `json:"resultXdr"`
	FeeMetaXDR   string `json:"feeMetaXdr"`
}

func newSendOutput(resp *soroban.SendTransactionResponse) sendOutput {
	return sendOutput{
		Hash:         resp.Hash,
		LatestLedger: resp.LatestLedger,
		FeeCharged:   resp.FeeCharged,
		MemoXDR:      resp.MemoXDR,
		SorobanMetaXDR: resp.SorobanMetaXDR,
		ResultXDR:    resp.ResultXDR,
		FeeMetaXDR:   resp.FeeMetaXDR,
	}
}

func printSend(w io.Writer, out sendOutput) {
	fmt.Fprintf(w, "hash:           %s\n", out.Hash)
	fmt.Fprintf(w, "latest ledger:  %d\n", out.LatestLedger)
	fmt.Fprintf(w, "fee charged:    %d stroops\n", out.FeeCharged)
	if out.MemoXDR != "" {
		fmt.Fprintf(w, "memo XDR:       %s\n", out.MemoXDR)
	}
	if out.SorobanMetaXDR != "" {
		fmt.Fprintf(w, "soroban meta XDR: %s\n", out.SorobanMetaXDR)
	}
	fmt.Fprintf(w, "result XDR:     %s\n", out.ResultXDR)
	fmt.Fprintf(w, "fee meta XDR:   %s\n", out.FeeMetaXDR)
}

// newSendCommand creates the sorobango send command.
func newSendCommand(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <transaction-envelope-xdr>",
		Short: "Submit a signed transaction to the Stellar network",
		Long: `Submit a signed transaction to the Stellar network for inclusion in a ledger.

The envelope is a base64 XDR TransactionEnvelope. Pass - to read it from
standard input instead, which is easier for the long envelopes real
transactions produce:

  sorobango send - < envelope.txt

The transaction must be fully signed before submission. A successful
submission only means the transaction was accepted for inclusion in a ledger -
it does not guarantee the transaction succeeded. To determine if the
transaction succeeded, examine the result XDR in the output.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envelope, err := readSendEnvelope(cmd, args[0])
			if err != nil {
				return err
			}

			client, err := opts.client()
			if err != nil {
				return err
			}

			resp, err := client.SendTransaction(cmd.Context(), envelope)
			if err != nil {
				return err
			}

			out := newSendOutput(resp)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printSend(w, out)
			})
		},
	}

	return cmd
}