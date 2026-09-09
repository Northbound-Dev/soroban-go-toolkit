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
