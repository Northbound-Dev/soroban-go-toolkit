<div style="text-align: justify">
# soroban-go-toolkit
🌐 https://soroban-go-toolkit.vercel.app/

[![CI](https://github.com/Northbound-Dev/soroban-go-toolkit/actions/workflows/ci.yml/badge.svg)](https://github.com/Northbound-Dev/soroban-go-toolkit/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Northbound-Dev/soroban-go-toolkit.svg)](https://pkg.go.dev/github.com/Northbound-Dev/soroban-go-toolkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A Go client library and CLI for reading Soroban smart contract state on the
Stellar network.

## Why this exists

Soroban tooling is concentrated in two ecosystems. Contract authors write Rust
against [`rs-soroban-sdk`](https://github.com/stellar/rs-soroban-sdk), and
frontends talk to the network through
[`js-stellar-sdk`](https://github.com/stellar/js-stellar-sdk) and
`soroban-react`.

Go is where a lot of backend infrastructure actually lives — indexers, bots,
monitoring, internal services — and Go developers reaching for Soroban have had
to hand-roll JSON-RPC calls and wrestle XDR themselves. The official
[`stellar/go`](https://github.com/stellar/go) monorepo provides excellent XDR
definitions and transaction building, but no dedicated Soroban RPC client.

This project fills that gap: typed request and response structs for the Stellar
RPC API, contract state reads that don't require you to build ledger keys by
hand, and a CLI for the same operations from a shell or a script.

## Requirements

Go 1.24 or later.

## Install

As a library:

```sh
go get github.com/Northbound-Dev/soroban-go-toolkit
```

As a CLI:

```sh
go install github.com/Northbound-Dev/soroban-go-toolkit/cmd/sorobango@latest
```

## Quickstart

This is a complete program. It runs against the public Stellar testnet with no
configuration and no account.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

func main() {
	// With no options the client targets the public Stellar testnet.
	client, err := soroban.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	health, err := client.GetHealth(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("endpoint status:", health.Status)

	ledger, err := client.GetLatestLedger(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("latest ledger:  %d (protocol %d)\n", ledger.Sequence, ledger.ProtocolVersion)
}
```

To run it from scratch:

```sh
mkdir soroban-demo && cd soroban-demo
go mod init soroban-demo
go get github.com/Northbound-Dev/soroban-go-toolkit
# save the program above as main.go
go run .
```

A runnable version, including an optional contract read, lives in
[`examples/quickstart`](examples/quickstart):

```sh
go run ./examples/quickstart
go run ./examples/quickstart -contract CXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

### Reading contract state

Contract storage is read through `getLedgerEntries`, which takes XDR ledger
keys. The client builds them for you:

```go
key, err := soroban.SymbolKey("COUNTER")
if err != nil {
	log.Fatal(err)
}

data, err := client.GetContractData(ctx, contractID, key, soroban.DurabilityPersistent)
switch {
case errors.Is(err, soroban.ErrEntryNotFound):
	log.Fatal("the contract has not written COUNTER yet")
case err != nil:
	log.Fatal(err)
}

fmt.Println("value type:", data.Value().Type)
```

`GetContractInstance` is the shortest way to check whether a contract is
deployed at an address at all — `ErrEntryNotFound` means nothing is there.

### Choosing a network

`soroban.TestnetURL` is the default and `soroban.FuturenetURL` is also defined.
There is no free Stellar-operated RPC endpoint for the public network, so
mainnet needs an instance you run yourself or obtain from a provider:

```go
client, err := soroban.New(
	soroban.WithURL("https://your-mainnet-rpc.example.com"),
	soroban.WithTimeout(10*time.Second),
)
```


### Detailed Usage Examples

#### Reading Different ScVal Types

The client provides helper functions to construct ledger keys for various types. To decode the returned `xdr.ScVal`, you can use the `github.com/stellar/go/xdr` package or perform a type switch on the `ScVal.Type` field.

```go
import (
	"context"
	"fmt"
	"log"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
	"github.com/stellar/go/xdr"
)

func readValues(ctx context.Context, client *soroban.Client, contractID string) {
	// Read a string value
	stringKey, err := soroban.SymbolKey("GREETING")
	if err != nil {
		log.Fatal(err)
	}
	stringData, err := client.GetContractData(ctx, contractID, stringKey, soroban.DurabilityPersistent)
	if err != nil {
		log.Fatal(err)
	}
	var str string
	xdr.Unmarshal(stringData.Value(), &str)
	fmt.Println("Greeting:", str)

	// Read an i32 integer
	i32Key, err := soroban.SymbolKey("COUNTER_I32")
	if err != nil {
		log.Fatal(err)
	}
	i32Data, err := client.GetContractData(ctx, contractID, i32Key, soroban.DurabilityPersistent)
	if err != nil {
		log.Fatal(err)
	}
	var i32 int32
	xdr.Unmarshal(i32Data.Value(), &i32)
	fmt.Println("Counter i32:", i32)

	// Read an i64 integer
	i64Key, err := soroban.SymbolKey("BIG_COUNTER")
	if err != nil {
		log.Fatal(err)
	}
	i64Data, err := client.GetContractData(ctx, contractID, i64Key, soroban.DurabilityPersistent)
	if err != nil {
		log.Fatal(err)
	}
	var i64 int64
	xdr.Unmarshal(i64Data.Value(), &i64)
	fmt.Println("Counter i64:", i64)

	// Read a vector (array) of strings
	vecKey, err := soroban.SymbolKey("ITEMS")
	if err != nil {
		log.Fatal(err)
	}
	vecData, err := client.GetContractData(ctx, contractID, vecKey, soroban.DurabilityPersistent)
	if err != nil {
		log.Fatal(err)
	}
	var vec []string
	xdr.Unmarshal(vecData.Value(), &vec)
	fmt.Println("Items:", vec)

	// Read a map from string to i32
	mapKey, err := soroban.SymbolKey("BALANCES")
	if err != nil {
		log.Fatal(err)
	}
	mapData, err := client.GetContractData(ctx, contractID, mapKey, soroban.DurabilityPersistent)
	if err != nil {
		log.Fatal(err)
	}
	var m map[string]int32
	xdr.Unmarshal(mapData.Value(), &m)
	fmt.Println("Balances:", m)
}
```

#### Simulating Contract Function Calls

To invoke a contract function (without submitting a transaction), use the `SimulateTransaction` method. This requires constructing a transaction with the appropriate invocations.

```go
import (
	"context"
	"fmt"
	"log"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
	"github.com/stellar/go/txnbuild"
)

func simulateContract(ctx context.Context, client *soroban.Client, contractID string) {
	// Build a transaction that invokes a contract function
	// Note: This example assumes you have a funded account for the source.
	// For simulation only, the source account does not need to be funded.
	sourceAccount := txnbuild.Account{
		Address: "GDG...", // replace with your account address
		Sequence: 1,
	}
	invoke := txnbuild.InvokeHostFunction{
		ContractAddress: contractID,
		FunctionName:    "increment",
		Args: []txnbuild.ScVal{
			txnbuild.ScVec{[]txnbuild.ScVal{
				txnbuild.ScvString("COUNTER"),
			}}.ToScVal(),
		},
	}
	tx, err := txnbuild.BuildTx(
		sourceAccount,
		txnbuild.Network{NetworkPassphrase: txnbuild.TestNetworkPassphrase},
		&invoke,
		txnbuild.BuildTxOpts{
			InheritMinimalFee: true,
			PreflightMemo:     true,
			// For simulation, we don't need to sign
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	txe, err := tx.Base64()
	if err != nil {
		log.Fatal(err)
	}

	// Simulate the transaction
	simResult, err := client.SimulateTransaction(ctx, txe)
	if err != nil {
		log.Fatal(err)
	}
	if simResult.Error != nil {
		fmt.Printf("Simulation error: %s\n", simResult.Error)
		return
	}
	fmt.Println("Simulation success:", simResult)
}
```

> **Note**: The above examples require the `github.com/stellar/go` package for transaction building and XDR unmarshalling. For pure reading examples, only the soroban-go-toolkit is needed.

See the [`examples/`](examples/) directory for runnable examples that you can adapt.

## Supported RPC methods

| Stellar RPC method | Client method | CLI command |
| --- | --- | --- |
| `getHealth` | `GetHealth` | `sorobango health` |
| `getLatestLedger` | `GetLatestLedger` | `sorobango latest-ledger` |
| `getLedgerEntries` | `GetLedgerEntries` | `sorobango ledger-entries` |
| `simulateTransaction` | `SimulateTransaction` | `sorobango simulate` |

Built on top of `getLedgerEntries`, for reading contract state without
constructing XDR by hand:

| Helper | CLI command | Purpose |
| --- | --- | --- |
| `GetContractData` | `sorobango contract-data` | Read one value from contract storage |
| `GetContractInstance` | `sorobango contract-instance` | Read a contract's instance entry |

Note that `getContractData` is not a Stellar RPC method — it existed in early
Soroban previews and was removed. Contract state is read via `getLedgerEntries`,
which is what these helpers do.

## CLI

Every command takes `--rpc-url`, `--timeout`, and `--json`, and exits non-zero
on failure so it composes in scripts.

```sh
sorobango health
sorobango latest-ledger --json
sorobango contract-data CXXX...  COUNTER --durability persistent
sorobango contract-instance CXXX...
sorobango simulate - < envelope.txt
```

Two failure cases are deliberately reflected in the exit code rather than only
in the output: `health` fails when the endpoint reports anything but healthy, so
it works as a readiness probe, and `simulate` fails when the contract call
itself fails even though the RPC call succeeded.

## Not yet implemented

These are tracked as issues and are good places to start contributing:

- `getEvents` — contract event queries
- `getTransaction` — transaction status and result lookup
- `sendTransaction` — transaction submission
- `getNetwork` and `getVersionInfo` — endpoint metadata
- Typed decoding for structured `ScVal` values (maps, vectors, `i128`/`u256`)


## Troubleshooting

### Common Issues

**Error: "command not found: sorobango" after installation**
This usually means your `$GOPATH/bin` directory is not in your system's PATH. After installing with `go install`, add `$GOPATH/bin` to your PATH:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

To make this permanent, add the above line to your shell profile (~/.bashrc, ~/.zshrc, etc.).

**Error: timeout or connection refused when connecting to Stellar RPC**
Ensure you have an active internet connection and that the RPC endpoint is accessible. You can test connectivity with:

```sh
curl -s https://soroban-testnet.stellar.org
```

If using a custom RPC endpoint, verify the URL is correct and the service is running.

**Error: contract not found when it should exist**
Double-check that you're using the correct network (testnet vs futurenet vs mainnet) and that the contract address is for that specific network. Contract addresses are network-specific.

### Getting Help
If you encounter issues not covered here:
1. Check the [existing issues](https://github.com/Northbound-Dev/soroban-go-toolkit/issues)
2. Create a new issue with detailed information about your problem
3. Include:
   - Your Go version (`go version`)
   - The soroban-go-toolkit version (from `go list -m github.com/Northbound-Dev/soroban-go-toolkit`)
   - The exact command you're running
   - The full error message
   - Steps to reproduce the issue
## Status

Pre-1.0. The implemented methods are tested and working, but the exported API
may still change before a v1.0 tag.

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for how to
run the tests, what's expected in a pull request, how issues are sized, and the
maintainer response-time commitment.

## License

MIT — see [LICENSE](LICENSE).
</div>
