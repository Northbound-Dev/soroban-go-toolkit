// Package jsonrpc implements the minimal subset of JSON-RPC 2.0 needed to talk
// to a Stellar RPC server: single, non-batched calls carried over HTTP POST.
//
// It is deliberately internal. The exported API in pkg/soroban translates these
// wire-level concerns into typed responses and exported error types, so that
// callers never have to reason about JSON-RPC framing.
package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// Version is the only JSON-RPC version Stellar RPC speaks.
const Version = "2.0"

// maxErrorBodyBytes bounds how much of a non-JSON error body is quoted back to
// the caller, so that a stray HTML error page from a proxy cannot produce a
// multi-megabyte error string.
const maxErrorBodyBytes = 512

// Error is a JSON-RPC error object. A non-nil Error means the server accepted
// the request and rejected the call; transport failures surface as HTTPError or
// as the underlying net/http error instead.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *Error) Error() string {
	if len(e.Data) > 0 {
		return fmt.Sprintf("jsonrpc: %s (code %d): %s", e.Message, e.Code, e.Data)
	}
	return fmt.Sprintf("jsonrpc: %s (code %d)", e.Message, e.Code)
}

// HTTPError reports a response whose HTTP status was outside the 2xx range,
// which means the request never reached the JSON-RPC layer.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("jsonrpc: http %s: %s", e.Status, e.Body)
	}
	return fmt.Sprintf("jsonrpc: http %s", e.Status)
}

type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Client performs JSON-RPC calls against a single endpoint. It is safe for
// concurrent use by multiple goroutines.
type Client struct {
	url    string
	http   *http.Client
	lastID atomic.Uint64
}

// NewClient returns a Client that posts to url using httpClient, which must not
// be nil.
func NewClient(url string, httpClient *http.Client) *Client {
	return &Client{url: url, http: httpClient}
}

// Call invokes method with params and decodes the result into out. A nil out
// discards the result; a nil params omits the member entirely, which is what
// parameterless methods such as getHealth expect.
func (c *Client) Call(ctx context.Context, method string, params, out any) error {
	id := c.lastID.Add(1)

	body, err := json.Marshal(request{JSONRPC: Version, ID: id, Method: method, Params: params})
	if err != nil {
		return fmt.Errorf("jsonrpc: encode %s request: %w", method, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("jsonrpc: build %s request: %w", method, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("jsonrpc: %s: %w", method, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(httpResp.Body, maxErrorBodyBytes))
		return &HTTPError{
			StatusCode: httpResp.StatusCode,
			Status:     httpResp.Status,
			Body:       string(bytes.TrimSpace(snippet)),
		}
	}

	var rpcResp response
	if err := json.NewDecoder(httpResp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("jsonrpc: decode %s response: %w", method, err)
	}

	// Checked before the id, because error responses are permitted to carry a
	// null id when the server could not parse the request that caused them.
	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	// A mismatched id means the payload belongs to some other request; treating
	// it as valid would silently hand the caller another call's data.
	if rpcResp.ID != id {
		return fmt.Errorf("jsonrpc: %s: response id %d does not match request id %d", method, rpcResp.ID, id)
	}

	if out == nil {
		return nil
	}
	if len(rpcResp.Result) == 0 {
		return fmt.Errorf("jsonrpc: %s: response carried neither result nor error", method)
	}
	if err := json.Unmarshal(rpcResp.Result, out); err != nil {
		return fmt.Errorf("jsonrpc: decode %s result: %w", method, err)
	}
	return nil
}
