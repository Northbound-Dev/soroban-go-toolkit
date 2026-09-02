package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stellar/go/xdr"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// parseAuthMode maps the flag value onto the client's AuthMode.
func parseAuthMode(value string) (soroban.AuthMode, error) {
	switch mode := soroban.AuthMode(strings.ToLower(value)); mode {
	case soroban.AuthModeEnforce, soroban.AuthModeRecord, soroban.AuthModeRecordAllowNonRoot:
		return mode, nil
	default:
		return "", fmt.Errorf("unknown auth mode %q, want enforce, record, or record_allow_nonroot", value)
	}
}

// readEnvelope resolves the envelope argument, treating "-" as standard input.
func readEnvelope(cmd *cobra.Command, arg string) (string, error) {
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

func printSimulation(w io.Writer, resp *soroban.SimulateTransactionResponse) {
	fmt.Fprintf(w, "latest ledger:    %d\n", resp.LatestLedger)

	if fee, err := resp.MinResourceFeeInt64(); err == nil {
		fmt.Fprintf(w, "min resource fee: %d stroops\n", fee)
	} else if resp.MinResourceFee != "" {
		fmt.Fprintf(w, "min resource fee: %s\n", resp.MinResourceFee)
	}

	if resp.Err() != nil {
		fmt.Fprintf(w, "\nsimulation failed: %s\n", resp.Error)
		// The diagnostic events are usually the only explanation of why a host
		// call failed, so they are worth printing even unparsed.
		if len(resp.Events) > 0 {
			fmt.Fprintf(w, "\ndiagnostic events (%d):\n", len(resp.Events))
			for _, event := range resp.Events {
				fmt.Fprintf(w, "  %s\n", event)
			}
		}
		return
	}

	for i, result := range resp.Results {
		fmt.Fprintf(w, "\nresult %d:\n", i)

		var value xdr.ScVal
		if err := xdr.SafeUnmarshalBase64(result.XDR, &value); err == nil {
			fmt.Fprintf(w, "  returns:      %s (%s)\n", formatScVal(value), value.Type)
		} else {
			fmt.Fprintf(w, "  returns:      %s\n", result.XDR)
		}
		if len(result.Auth) > 0 {
			fmt.Fprintf(w, "  auth entries: %d\n", len(result.Auth))
		}
	}

	if len(resp.StateChanges) > 0 {
		fmt.Fprintf(w, "\nstate changes (%d):\n", len(resp.StateChanges))
		for _, change := range resp.StateChanges {
			fmt.Fprintf(w, "  %-8s %s\n", change.Type, change.Key)
		}
	}

	if resp.NeedsRestore() {
		fmt.Fprintf(w, "\narchived entries must be restored before submitting")
		fmt.Fprintf(w, " (restore fee %s stroops)\n", resp.RestorePreamble.MinResourceFee)
	}
}

func newSimulateCommand(opts *options) *cobra.Command {
	var (
		instructionLeeway uint64
		authMode          string
	)

	cmd := &cobra.Command{
		Use:   "simulate <transaction-envelope-xdr>",
		Short: "Simulate a transaction without submitting it",
		Long: `Run a transaction against the current ledger without submitting it, reporting
the resource fee, return value, and any authorization entries it needs.

The envelope is a base64 XDR TransactionEnvelope. Pass - to read it from
standard input instead, which is easier for the long envelopes real
transactions produce:

  sorobango simulate - < envelope.txt

Simulation writes nothing and the transaction does not need to be signed. A
transaction whose contract call fails still returns a valid response; the
command reports the failure and exits non-zero.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envelope, err := readEnvelope(cmd, args[0])
			if err != nil {
				return err
			}

			var simulateOpts []soroban.SimulateOption
			if instructionLeeway > 0 {
				simulateOpts = append(simulateOpts, soroban.WithInstructionLeeway(instructionLeeway))
			}
			if authMode != "" {
				mode, err := parseAuthMode(authMode)
				if err != nil {
					return err
				}
				simulateOpts = append(simulateOpts, soroban.WithAuthMode(mode))
			}

			client, err := opts.client()
			if err != nil {
				return err
			}

			resp, err := client.SimulateTransaction(cmd.Context(), envelope, simulateOpts...)
			if err != nil {
				return err
			}

			if err := emit(cmd, opts.asJSON, resp, func(w io.Writer) {
				printSimulation(w, resp)
			}); err != nil {
				return err
			}

			// Surfaced through the exit code as well as the output, so a script
			// cannot mistake a failed contract call for success.
			return resp.Err()
		},
	}

	flags := cmd.Flags()
	flags.Uint64Var(&instructionLeeway, "instruction-leeway", 0,
		"extra CPU instructions to allow beyond the simulated cost")
	flags.StringVar(&authMode, "auth-mode", "",
		"authorization mode: enforce, record, or record_allow_nonroot")

	return cmd
}
