package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stellar/go/xdr"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// contractDataOutput is the CLI's own JSON shape for a contract storage read.
//
// The client returns XDR types, which would serialise into something verbose
// and tied to the SDK's internal representation. Declaring the output
// explicitly keeps --json a stable contract for scripts.
type contractDataOutput struct {
	ContractID            string  `json:"contractId"`
	Durability            string  `json:"durability"`
	ValueType             string  `json:"valueType"`
	Value                 string  `json:"value"`
	ValueXDR              string  `json:"valueXdr"`
	LastModifiedLedgerSeq uint32  `json:"lastModifiedLedgerSeq"`
	LiveUntilLedgerSeq    *uint32 `json:"liveUntilLedgerSeq,omitempty"`
	LatestLedger          uint32  `json:"latestLedger"`
}

func newContractDataOutput(contractID, durability string, data *soroban.ContractData) contractDataOutput {
	value := data.Value()

	encoded, err := xdr.MarshalBase64(value)
	if err != nil {
		encoded = ""
	}

	return contractDataOutput{
		ContractID:            contractID,
		Durability:            durability,
		ValueType:             value.Type.String(),
		Value:                 formatScVal(value),
		ValueXDR:              encoded,
		LastModifiedLedgerSeq: data.LastModifiedLedgerSeq,
		LiveUntilLedgerSeq:    data.LiveUntilLedgerSeq,
		LatestLedger:          data.LatestLedger,
	}
}

func printContractData(w io.Writer, out contractDataOutput) {
	fmt.Fprintf(w, "contract:      %s\n", out.ContractID)
	fmt.Fprintf(w, "durability:    %s\n", out.Durability)
	fmt.Fprintf(w, "value:         %s\n", out.Value)
	fmt.Fprintf(w, "value type:    %s\n", out.ValueType)
	fmt.Fprintf(w, "last modified: ledger %d\n", out.LastModifiedLedgerSeq)
	if out.LiveUntilLedgerSeq != nil {
		fmt.Fprintf(w, "live until:    ledger %d\n", *out.LiveUntilLedgerSeq)
	}
	fmt.Fprintf(w, "latest ledger: %d\n", out.LatestLedger)
}

// parseDurability maps the flag value onto the XDR enum.
func parseDurability(value string) (xdr.ContractDataDurability, error) {
	switch strings.ToLower(value) {
	case "persistent":
		return soroban.DurabilityPersistent, nil
	case "temporary":
		return soroban.DurabilityTemporary, nil
	default:
		return 0, fmt.Errorf("unknown durability %q, want persistent or temporary", value)
	}
}

func newContractDataCommand(opts *options) *cobra.Command {
	var durability string

	cmd := &cobra.Command{
		Use:   "contract-data <contract-id> <key>",
		Short: "Read a value from a contract's storage",
		Long: `Read a value from a contract's storage by its symbol key.

The contract id is a strkey contract address beginning with C. The key is the
symbol the contract stores the value under, for example COUNTER.

Persistent and temporary storage are separate areas, so a value written to one
is not visible in the other; --durability selects which to read.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			contractID, symbol := args[0], args[1]

			parsedDurability, err := parseDurability(durability)
			if err != nil {
				return err
			}

			key, err := soroban.SymbolKey(symbol)
			if err != nil {
				return err
			}

			client, err := opts.client()
			if err != nil {
				return err
			}

			data, err := client.GetContractData(cmd.Context(), contractID, key, parsedDurability)
			if err != nil {
				return err
			}

			out := newContractDataOutput(contractID, strings.ToLower(durability), data)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printContractData(w, out)
			})
		},
	}

	cmd.Flags().StringVar(&durability, "durability", "persistent",
		"storage area to read: persistent or temporary")

	return cmd
}

func newContractInstanceCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "contract-instance <contract-id>",
		Short: "Read a contract's instance entry",
		Long: `Read a contract's instance entry, which holds its executable reference and
instance storage.

This is the cheapest way to confirm a contract is actually deployed at an
address, since a missing instance entry means nothing is there.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			contractID := args[0]

			client, err := opts.client()
			if err != nil {
				return err
			}

			data, err := client.GetContractInstance(cmd.Context(), contractID)
			if err != nil {
				return err
			}

			out := newContractDataOutput(contractID, "persistent", data)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printContractData(w, out)
			})
		},
	}
}
