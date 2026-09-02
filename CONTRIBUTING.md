# Contributing to soroban-go-toolkit

Thanks for your interest in contributing. This project exists to make Soroban
usable from Go, and most of what's missing is well-scoped work that a new
contributor can pick up without deep Stellar knowledge.

## Getting set up

You need Go 1.24 or later. Nothing else — no Stellar account, no local network,
no API keys.

```sh
git clone https://github.com/Northbound-Dev/soroban-go-toolkit.git
cd soroban-go-toolkit
go build ./...
go test ./...
```

## Running the tests

```sh
go test ./...                 # the whole suite
go test -race ./...           # what CI runs
go test ./pkg/soroban -run TestGetHealth -v
go test ./... -cover
```

**Tests must not touch the network.** The HTTP layer is mocked with
`net/http/httptest` throughout, so the suite is fast, deterministic, and works
offline. A test that reaches a live endpoint will be flaky in CI and won't be
merged.

The one exception is `examples/quickstart`, which deliberately calls live
testnet. CI compiles it but never runs it. To try it yourself:

```sh
go run ./examples/quickstart
```

## Coding style

- **`gofmt` is mandatory.** CI fails on unformatted code. Run `gofmt -w .`
  before committing.
- **`go vet ./...` and `golangci-lint` must pass.** CI runs golangci-lint with
  its default linters (`errcheck`, `govet`, `ineffassign`, `staticcheck`,
  `unused`). There's no checked-in config, so local and CI behaviour match.
- **Write idiomatic Go.** Wrap errors with `fmt.Errorf("...: %w", err)` so
  callers can use `errors.Is`/`errors.As`. Return concrete types, not
  `map[string]any`. Don't add abstraction layers that only have one
  implementation.
- **Comment the why, not the what.** Explain a non-obvious decision, a protocol
  quirk, or a footgun. Don't narrate code that already reads clearly.
- **New RPC methods need typed request and response structs**, matching the
  field names the API actually sends. Where a value arrives as base64 XDR, name
  the field so the XDR type it decodes to is unambiguous.

## How issues are sized

Issues carry a complexity label that maps to Drips Wave points:

| Label | Points | What it means |
| --- | --- | --- |
| **Trivial** | 100 | Typos, small bug fixes, minor copy or doc changes |
| **Medium** | 150 | Standard features, or bug fixes needing real investigation |
| **High** | 200 | Complex features, refactors, new integrations |

Sizing reflects the actual work involved, not how much we want it done. If you
start an issue and find it's substantially bigger or smaller than labelled, say
so on the issue — it will be re-labelled rather than left wrong.

Issues labelled **good first issue** are self-contained and have an existing
pattern in the codebase to follow. Each one names the files you'll likely touch.

## Claiming an issue

Comment on the issue to claim it, and it'll be assigned to you. Please don't
open a pull request for an issue assigned to someone else.

If you've claimed something and your plans change, just say so on the issue.
That's genuinely fine and much better than an issue sitting assigned and
untouched — it frees it up for someone else.

## Pull requests

A good pull request here:

- **Does one thing.** A new RPC method plus an unrelated refactor should be two
  pull requests.
- **Includes tests.** New client methods need tests covering the request that
  gets sent, successful decoding, and at least one failure path. Table-driven
  where there are several similar cases.
- **Passes CI before review.** Build, `go vet`, `go test -race`, `gofmt`,
  `go mod tidy`, and lint. You can run all of it locally.
- **Explains the why in the description.** Link the issue it closes. If you made
  a judgement call — a name, a field type, an error strategy — say so, so review
  can focus there.
- **Updates the docs it affects.** A new RPC method belongs in the README's
  supported-methods table and, if it gets a CLI command, in the CLI section.
- **Leaves no commented-out code or `TODO`s.** If something is deliberately out
  of scope, open a follow-up issue and link it.

Draft pull requests are welcome if you want feedback on an approach before
finishing. Ask questions on the issue rather than guessing — it's faster for
everyone.

## Maintainer commitments

- **Issue assignment: within 24–48 hours during an active Drips Wave.** Getting
  blocked waiting for an assignment is the worst part of contributing to a
  Wave project, so this is the commitment taken most seriously.
- **Pull request review: within 24–48 hours during an active Wave.** If a review
  will take longer, you'll get a comment saying so and when to expect it, rather
  than silence.
- **Review feedback will be specific and actionable.** If something needs
  changing, you'll be told what and why, not just that it isn't right.
- **Outside an active Wave, expect a few days.** Still acknowledged, just not on
  the same clock.

If a pull request or issue has gone quiet past these windows, a comment tagging
the maintainer is welcome and not considered rude.

## Reporting bugs

Open an issue with the Go version, the RPC endpoint you were hitting, what you
expected, what happened, and the smallest snippet that reproduces it. If it
involves a specific contract or ledger entry, include the address — being able
to reproduce a read against real data makes a large difference.

## Security

Don't open a public issue for a security problem. Report it privately through
the repository's security advisory page instead.
