package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// networkOutput is the CLI's own JSON shape for network info.
//
// The client returns types that would serialise into something verbose
// and tied to the SDK's internal representation. Declaring the output
// explicitly keeps --json a stable contract for scripts.
type networkOutput struct {
	Passport        string `json:"passport"`
	ProtocolVersion uint32 `json:"protocolVersion"`
}

func newNetworkOutput(resp *soroban.NetworkResponse) networkOutput {
	return networkOutput{
		Passport:        resp.Passport,
		ProtocolVersion: resp.ProtocolVersion,
	}
}

func printNetwork(w io.Writer, out networkOutput) {
	fmt.Fprintf(w, "passport:         %s\n", out.Passport)
	fmt.Fprintf(w, "protocol version: %d\n", out.ProtocolVersion)
}

// newNetworkCommand creates the sorobango network command.
func newNetworkCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "network",
		Short: "Show information about the connected Stellar network",
		Long: `Display information about the Stellar network the RPC endpoint is connected to.

This includes the network passport (identifier) and the protocol version in use.
This is useful for verifying you're connected to the correct network (testnet,
futurenet, mainnet, etc.) and for debugging connectivity issues.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			resp, err := client.GetNetwork(cmd.Context())
			if err != nil {
				return err
			}

			out := newNetworkOutput(resp)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printNetwork(w, out)
			})
		},
	}
}

// versionInfoOutput is the CLI's own JSON shape for version info.
//
// The client returns types that would serialise into something verbose
// and tied to the SDK's internal representation. Declaring the output
// explicitly keeps --json a stable contract for scripts.
type versionInfoOutput struct {
	CoreVersion        string `json:"coreVersion"`
	StellarCoreVersion string `json:"stellarCoreVersion"`
}

func newVersionInfoOutput(resp *soroban.VersionInfoResponse) versionInfoOutput {
	return versionInfoOutput{
		CoreVersion:        resp.CoreVersion,
		StellarCoreVersion: resp.StellarCoreVersion,
	}
}

func printVersionInfo(w io.Writer, out versionInfoOutput) {
	fmt.Fprintf(w, "core version:          %s\n", out.CoreVersion)
	fmt.Fprintf(w, "stellar-core version:  %s\n", out.StellarCoreVersion)
}

// newVersionInfoCommand creates the sorobango version command.
func newVersionInfoCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information for the connected Stellar core and horizon",
		Long: `Display version information about the Stellar core and horizon instances
the RPC endpoint is connected to.

This includes the horizon core version and the stellar-core version, which is
useful for debugging compatibility issues and verifying feature availability.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			resp, err := client.GetVersionInfo(cmd.Context())
			if err != nil {
				return err
			}

			out := newVersionInfoOutput(resp)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printVersionInfo(w, out)
			})
		},
	}
}
