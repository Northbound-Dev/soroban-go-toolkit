// Command sorobango is a command line client for the Stellar RPC API.
//
// It wraps the pkg/soroban client, exposing each RPC method as a subcommand.
// Every command prints human-readable output by default and raw JSON with
// --json, and exits non-zero when a call fails, so it composes in scripts.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Cancelling on interrupt means an in-flight request is abandoned promptly
	// rather than hanging until the timeout expires.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRootCommand().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "sorobango:", err)
		os.Exit(1)
	}
}
