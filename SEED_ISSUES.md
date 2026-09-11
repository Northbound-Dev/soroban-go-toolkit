## Seed Issues for soroban-go-toolkit

### Title: Implement getEvents RPC method
**Description:** Add support for the `getEvents` RPC method to query contract events on the Stellar network. This would allow users to listen for and retrieve events emitted by smart contracts.
**Acceptance Criteria:**
- [ ] Add `GetEvents` method to the `pkg/soroban` client that calls the `getEvents` RPC endpoint
- [ ] Define appropriate request and response structs matching the Stellar RPC API specification
- [ ] Add unit tests with mocked HTTP layer
- [ ] Add CLI command `sorobango events` in `cmd/sorobango`
- [ ] Update README.md to document the new method in the supported methods table
**Suggested Complexity:** Medium (150 Points)

### Title: Implement getTransaction RPC method
**Description:** Add support for the `getTransaction` RPC method to look up the status and result of a submitted transaction by its hash.
**Acceptance Criteria:**
- [ ] Add `GetTransaction` method to the `pkg/soroban` client that calls the `getTransaction` RPC endpoint
- [ ] Define appropriate request and response structs matching the Stellar RPC API specification
- [ ] Add unit tests with mocked HTTP layer
- [ ] Add CLI command `sorobango tx` or `sorobango transaction` in `cmd/sorobango`
- [ ] Update README.md to document the new method in the supported methods table
**Suggested Complexity:** Medium (150 Points)

### Title: Implement sendTransaction RPC method
**Description:** Add support for the `sendTransaction` RPC method to submit a signed transaction to the Stellar network for processing.
**Acceptance Criteria:**
- [ ] Add `SendTransaction` method to the `pkg/soroban` client that calls the `sendTransaction` RPC endpoint
- [ ] Define appropriate request and response structs matching the Stellar RPC API specification
- [ ] Add unit tests with mocked HTTP layer
- [ ] Add CLI command `sorobango send` in `cmd/sorobango`
- [ ] Update README.md to document the new method in the supported methods table
**Suggested Complexity:** High (200 Points)

### Title: Add getNetwork and getVersionInfo RPC methods
**Description:** Add support for the `getNetwork` and `getVersionInfo` RPC methods to retrieve endpoint metadata such as network passphrase and protocol version.
**Acceptance Criteria:**
- [ ] Add `GetNetwork` and `GetVersionInfo` methods to the `pkg/soroban` client
- [ ] Define appropriate request and response structs matching the Stellar RPC API specification
- [ ] Add unit tests with mocked HTTP layer
- [ ] Add CLI commands `sorobango network` and `sorobango version` in `cmd/sorobango`
- [ ] Update README.md to document the new methods in the supported methods table
**Suggested Complexity:** Trivial (100 Points)

### Title: Implement typed ScVal decoding helpers
**Description:** Add helper functions to decode common `ScVal` types (like maps, vectors, i128/u256) into native Go types for easier contract state interpretation.
**Acceptance Criteria:**
- [ ] Add functions to decode `ScVal` types into `map[string]interface{}`, `[]interface{}`, `*big.Int`, etc.
- [ ] Handle all `ScVal` types: void, bool, i32, i64, time, u32, u64, i128, u256, bytes, string, vec, map
- [ ] Add comprehensive unit tests
- [ ] Update documentation and examples to show usage
**Suggested Complexity:** High (200 Points)


### Title: Implement getNetwork RPC method
**Description:** Add support for the `getNetwork` RPC method to retrieve endpoint metadata such as network passphrase.
**Acceptance Criteria:**
- [ ] Add `GetNetwork` method to the `pkg/soroban` client that calls the `getNetwork` RPC endpoint
- [ ] Define appropriate request and response structs matching the Stellar RPC API specification
- [ ] Add unit tests with mocked HTTP layer
- [ ] Add CLI command `sorobango network` in `cmd/sorobango`
- [ ] Update README.md to document the new method in the supported methods table
**Suggested Complexity:** Trivial (100 Points)

### Title: Implement getVersionInfo RPC method
**Description:** Add support for the `getVersionInfo` RPC method to retrieve endpoint metadata such as protocol version and core version.
**Acceptance Criteria:**
- [ ] Add `GetVersionInfo` method to the `pkg/soroban` client that calls the `getVersionInfo` RPC endpoint
- [ ] Define appropriate request and response structs matching the Stellar RPC API specification
- [ ] Add unit tests with mocked HTTP layer
- [ ] Add CLI command `sorobango version` in `cmd/sorobango`
- [ ] Update README.md to document the new method in the supported methods table
**Suggested Complexity:** Trivial (100 Points)

### Title: Implement typed ScVal decoding helpers
**Description:** Add helper functions to decode common `ScVal` types (like maps, vectors, i128/u256) into native Go types for easier contract state interpretation.
**Acceptance Criteria:**
- [ ] Add functions to decode `ScVal` types into `map[string]interface{}`, `[]interface{}`, `*big.Int`, etc.
- [ ] Handle all `ScVal` types: void, bool, i32, i64, time, u32, u64, i128, u256, bytes, string, vec, map
- [ ] Add comprehensive unit tests
- [ ] Update documentation and examples to show usage
**Suggested Complexity:** High (200 Points)

### Title: Add retry mechanism for RPC calls
**Description:** Add configurable retry mechanism with exponential backoff to handle transient network failures when making RPC calls to the Stellar network.
**Acceptance Criteria:**
- [ ] Add optional retry configuration to `Client` via functional options
- [ ] Implement exponential backoff with jitter for retry delays
- [ ] Make retry behavior configurable (max attempts, base delay, max delay, etc.)
- [ ] Apply retry logic to idempotent RPC methods (getHealth, getLatestLedger, getLedgerEntries, etc.)
- [ ] Add unit tests for retry logic including success after retries and failure after max attempts
- [ ] Add documentation showing how to configure retry behavior
**Suggested Complexity:** Medium (150 Points)

### Title: Improve error handling with custom error types
**Description:** Replace generic error returns with typed errors that allow callers to distinguish between different failure scenarios (not found, timeout, rate limit, etc.).
**Acceptance Criteria:**
- [ ] Define custom error types for different failure scenarios (ErrNotFound, ErrTimeout, ErrRateLimit, etc.)
- [ ] Update existing methods to return typed errors where appropriate based on HTTP status codes and RPC error responses
- [ ] Add helper functions to check error types (IsNotFound(err), IsTimeout(err), IsRateLimit(err), etc.)
- [ ] Preserve backward compatibility by ensuring errors still work with existing error checking patterns
- [ ] Add unit tests for error type checking and error propagation
- [ ] Update documentation to document the new error types and how to use them
**Suggested Complexity:** Medium (150 Points)
