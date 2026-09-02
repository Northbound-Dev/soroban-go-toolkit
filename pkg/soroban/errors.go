package soroban

import (
	"errors"
	"fmt"

	"github.com/Northbound-Dev/soroban-go-toolkit/internal/jsonrpc"
)

// RPCError is returned when the server accepts a request and rejects the call,
// carrying the JSON-RPC error object verbatim. Code follows the JSON-RPC
// convention: -32601 for an unsupported method, -32602 for invalid params.
type RPCError struct {
	Code    int
	Message string
	// Data holds the error object's optional data member as raw JSON, empty
	// when the server sent none.
	Data string
}

func (e *RPCError) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("soroban: rpc error %d: %s: %s", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("soroban: rpc error %d: %s", e.Code, e.Message)
}

// HTTPError is returned when the endpoint replies with a non-2xx status, which
// means the request never reached the JSON-RPC layer — a wrong URL, a proxy
// error, or a rate limit rather than a rejected call.
type HTTPError struct {
	StatusCode int
	Status     string
	// Body holds the start of the response body, truncated to keep error
	// strings bounded.
	Body string
}

func (e *HTTPError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("soroban: http %s: %s", e.Status, e.Body)
	}
	return fmt.Sprintf("soroban: http %s", e.Status)
}

// SimulationError reports a simulation the server executed successfully but
// which failed inside the host, for example a panicking contract call. It is
// produced by SimulateTransactionResponse.Err rather than returned from
// SimulateTransaction, because the RPC call itself succeeded.
type SimulationError struct {
	Message string
	// Events holds the diagnostic events recorded during the failed
	// simulation, as base64 XDR DiagnosticEvent values. These usually explain
	// the failure and are worth logging.
	Events []string
}

func (e *SimulationError) Error() string {
	return fmt.Sprintf("soroban: simulation failed: %s", e.Message)
}

// translateError rewrites internal transport errors as this package's exported
// types, preserving anything it does not recognise.
func translateError(err error) error {
	if err == nil {
		return nil
	}

	var rpcErr *jsonrpc.Error
	if errors.As(err, &rpcErr) {
		return &RPCError{Code: rpcErr.Code, Message: rpcErr.Message, Data: string(rpcErr.Data)}
	}

	var httpErr *jsonrpc.HTTPError
	if errors.As(err, &httpErr) {
		return &HTTPError{StatusCode: httpErr.StatusCode, Status: httpErr.Status, Body: httpErr.Body}
	}

	return err
}
