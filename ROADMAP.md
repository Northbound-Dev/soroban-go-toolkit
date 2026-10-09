# Project Roadmap: soroban-go-toolkit

This roadmap outlines the past achievements, current sprint objectives, and future direction for `soroban-go-toolkit`. We use this document to coordinate contributor efforts, plan releases, and scope tasks for the **Drips Wave** program.

---

## Vision & Core Tenets

1. **Idiomatic Go First**: Deliver clean, robust APIs with strong type safety, context cancellation, native Go types, and zero hidden magic.
2. **Deterministic & Offline Testable**: Every client method and utility must be testable offline with mocked transports (`httptest`). No flaky CI runs depending on live testnets.
3. **Lossless XDR Interpretation**: Provide full-fidelity conversions between Stellar's binary XDR format, raw JSON-RPC structures, and typed Go structs.
4. **Developer Empowerment**: Offer both an embeddable Go library for infrastructure/bots and a capable CLI (`sorobango`) for shell scripts, devops, and rapid contract exploration.

---

## Milestones Overview

```
[Phase 1: Core Protocol Coverage]  ==>  [Phase 2: Developer Ergonomics]  ==>  [Phase 3: High-Level Utilities]  ==>  [Phase 4: v1.0 Production]
            (Completed)                         (Current Sprint)                       (Upcoming)                           (Planned)
```

---

## Phase 1: Core Protocol & RPC Coverage (Completed)

- [x] **Full Stellar RPC 21+ Method Support**:
  - `getHealth` — endpoint connectivity & readiness probe
  - `getLatestLedger` — sequence, close time, protocol version
  - `getLedgerEntries` — fetch raw ledger keys & storage entries
  - `getNetwork` — passphrase & protocol version inspection
  - `getVersionInfo` — Stellar RPC node version and commit metadata
  - `simulateTransaction` — contract execution simulation and resource profiling
  - `sendTransaction` — async transaction submission with status handling
  - `getTransaction` — transaction status (`SUCCESS`, `FAILED`, `PENDING`, `NOT_FOUND`) and execution results
  - `getEvents` — contract event querying with topic filters and pagination
  - `getFeeStats` — network fee distribution (p10..p99) and inclusion stats
  - `getTransactions` — historical ledger transaction range queries with pagination
- [x] **Contract State Helpers**:
  - `GetContractData` — targeted storage key lookup (`persistent`, `temporary`, `instance`)
  - `GetContractInstance` — contract executable, wasm hash, and instance state inspection
- [x] **Lossless Typed ScVal Decoding Engine**:
  - Primitives: `DecodeBool`, `DecodeU32`, `DecodeI32`, `DecodeU64`, `DecodeI64`
  - High-precision big integers: `DecodeU128`, `DecodeI128`, `DecodeU256`, `DecodeI256` using exact two's-complement `*big.Int` arithmetic
  - Strings, symbols, bytes, and address decoding
  - Complex nested structures: vectors, ordered maps, and recursive `DecodeNative`
- [x] **Transaction Orchestration**:
  - `SendAndAwaitTransaction` helper with configurable polling intervals and context deadlines
- [x] **Custom Headers & Authentication**:
  - Per-client and per-request custom HTTP headers (e.g., Bearer tokens for authenticated RPC providers like QuickNode, Blockdaemon)
- [x] **CLI Tool (`sorobango`)**:
  - Full CLI subcommand coverage matching all library methods
  - Pipe-friendly stdin support (`-`) and clean JSON formatting (`--json`)

---

## Phase 2: Developer Ergonomics & Contributor Infrastructure (Current Sprint)

Focuses on lowering onboarding friction, providing automated release tooling, and establishing standardized community contribution workflows.

- [x] **Repository Restructuring**:
  - Dedicated landing page in `website/` with isolated deployment configuration
  - Removal of obsolete seed files; centralized tracking in `ROADMAP.md`
- [x] **Standardized Issue & PR Infrastructure**:
  - Drips Wave contributor task templates (`.github/ISSUE_TEMPLATE/wave_task.md`) with explicit complexity sizing and acceptance criteria
  - Structured PR template (`.github/pull_request_template.md`) enforcing test coverage and lint verification
  - Automation scripts via `Makefile` (`make test`, `make lint`, `make cover`, `make build`)
- [x] **Release Engineering**:
  - Multi-platform automated releases via GoReleaser (Linux, macOS, Windows; AMD64/ARM64)
- [ ] **Wave Contributor Tasks**:
  - [ ] Bidirectional `ScVal` encoders (Native Go -> `xdr.ScVal`) (#WaveTask)
  - [ ] Transaction simulation resource fee estimator & footprint analyzer (#WaveTask)
  - [ ] Soroban contract event streaming & subscriber channels (#WaveTask)
  - [ ] Comprehensive runnable example suites for developers (#WaveTask)

---

## Phase 3: High-Level Contract Utilities & Streaming (Upcoming)

Focuses on higher-level abstractions that accelerate building bots, indexers, and off-chain services.

- [ ] **Event Streaming & Subscription Engine**:
  - Resilient `EventSubscriber` with exponential backoff and cursor bookmarking
  - Channel-based Go streaming (`<-chan ContractEvent`) with context cancellation
  - CLI `sorobango events --watch` command for terminal-based event tailing
- [ ] **Contract Invocation & Authorization Builder**:
  - Helpers for constructing `xdr.SorobanAuthorizationEntry` and address credentials
  - Simplified multi-signature and account auth invocation chaining
- [ ] **State Archival & TTL Management**:
  - Protocol 20/21/22 TTL inspection helpers on `LedgerEntryResult`
  - Detection of archived entries and automatic recommendation for `RestoreFootprintOp` or `BumpFootprintInstance`
- [ ] **Mock RPC Server (`pkg/sorobantest`)**:
  - In-memory mock Soroban RPC server for offline integration tests of third-party Go applications

---

## Phase 4: Production Hardening & v1.0.0 (Planned)

- [ ] **High-Throughput Batch Queries**:
  - HTTP connection pooling optimizations and multi-key JSON-RPC batching
- [ ] **Interactive CLI REPL**:
  - Shell mode (`sorobango repl`) for querying contracts, decoding storage, and inspecting transactions interactively
- [ ] **Type-Safe Contract Code Generation (Experimental)**:
  - CLI tool to generate type-safe Go bindings from contract WASM or interface specs

---

## How to Propose or Claim Work

If you are participating via **Drips Wave** or the broader open-source community:
1. Browse issues tagged [`wave-task`](https://github.com/Northbound-Dev/soroban-go-toolkit/issues?q=is%3Aissue+is%3Aopen+label%3Awave-task) or [`good first issue`](https://github.com/Northbound-Dev/soroban-go-toolkit/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).
2. Review our [CONTRIBUTING.md](CONTRIBUTING.md) guide for PR requirements and maintainer commitments.
3. Comment on the issue to get assigned before beginning work.
