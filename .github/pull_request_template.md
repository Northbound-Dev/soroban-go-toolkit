## Description

<!-- Briefly describe the changes introduced in this pull request and the rationale behind them. -->

## Related Issue / Wave Task

- Fixes / Addresses: #<!-- issue number -->
- Drips Wave Complexity: <!-- Trivial (100 pts) / Medium (150 pts) / High (200 pts) / N/A -->

## Type of Change

- [ ] New RPC method / client API feature
- [ ] New CLI command or flag
- [ ] Enhancement to existing functionality
- [ ] Bug fix
- [ ] Documentation update
- [ ] Refactor or test suite improvement

## Checklist

- [ ] My code follows the repository's idiomatic Go style and conventions.
- [ ] I have formatted my code using `gofmt -w .`.
- [ ] `go vet ./...` and `golangci-lint run` report no warnings or errors.
- [ ] I have added unit tests with mocked transports (`httptest`) covering success and failure paths.
- [ ] **Tests do not make live network calls.**
- [ ] All tests pass with race detector: `go test -race ./...`.
- [ ] I have updated the relevant documentation ([`README.md`](README.md), [`ROADMAP.md`](ROADMAP.md), or code comments).

## Testing Evidence

<!-- Paste the output of `go test -race ./...` or relevant test execution snippets below: -->

```sh
go test -v -race ./...
```
