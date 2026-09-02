package soroban

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// capturedRequest is the JSON-RPC envelope as it arrived at the test server,
// letting tests assert on what the client actually sent.
type capturedRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// serveResult returns a handler that decodes the incoming request into capture
// (when non-nil) and replies with result as the JSON-RPC result member, echoing
// the request id so that the client's id check passes.
func serveResult(t *testing.T, result string, capture *capturedRequest) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var req capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("server could not decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if capture != nil {
			*capture = req
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":%s}`, req.ID, result)
	}
}

// newTestClient points a Client at an httptest server running fn.
func newTestClient(t *testing.T, fn http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(fn)
	t.Cleanup(srv.Close)

	client, err := New(WithURL(srv.URL))
	if err != nil {
		t.Fatalf("New(WithURL(%q)) returned error: %v", srv.URL, err)
	}
	return client
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantURL string
		wantErr bool
	}{
		{
			name:    "defaults to testnet",
			wantURL: TestnetURL,
		},
		{
			name:    "explicit url",
			opts:    []Option{WithURL("https://rpc.example.com")},
			wantURL: "https://rpc.example.com",
		},
		{
			name:    "plain http is allowed for local nodes",
			opts:    []Option{WithURL("http://localhost:8000")},
			wantURL: "http://localhost:8000",
		},
		{
			name:    "empty url",
			opts:    []Option{WithURL("")},
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			opts:    []Option{WithURL("ws://rpc.example.com")},
			wantErr: true,
		},
		{
			name:    "missing host",
			opts:    []Option{WithURL("https://")},
			wantErr: true,
		},
		{
			name:    "relative url has no host",
			opts:    []Option{WithURL("rpc.example.com")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.opts...)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("New() = %v, want error", client.Endpoint())
				}
				return
			}
			if err != nil {
				t.Fatalf("New() returned unexpected error: %v", err)
			}
			if got := client.Endpoint(); got != tt.wantURL {
				t.Errorf("Endpoint() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

// recordingTransport answers every request from a fixed body, noting that it was
// used so that a test can prove which http.Client did the work.
type recordingTransport struct {
	calls int
	body  string
}

func (rt *recordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.calls++

	var req capturedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":%s}`, req.ID, rt.body)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}, nil
}

func TestWithHTTPClientIsUsed(t *testing.T) {
	transport := &recordingTransport{body: `{"status":"healthy"}`}
	custom := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	// The timeout option is deliberately also supplied: WithHTTPClient must win,
	// since the supplied client carries its own.
	client, err := New(WithTimeout(time.Minute), WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	health, err := client.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth() returned error: %v", err)
	}
	if !health.Healthy() {
		t.Error("Healthy() = false, want true")
	}
	if transport.calls != 1 {
		t.Errorf("custom transport handled %d calls, want 1", transport.calls)
	}
}

func TestCallSendsWellFormedRequest(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		captured       capturedRequest
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		serveResult(t, `{"status":"healthy"}`, &captured)(w, r)
	})

	if _, err := client.GetHealth(context.Background()); err != nil {
		t.Fatalf("GetHealth() returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("HTTP method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if captured.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want 2.0", captured.JSONRPC)
	}
	if captured.Method != "getHealth" {
		t.Errorf("method = %q, want getHealth", captured.Method)
	}
	// getHealth takes no parameters, so the member must be omitted rather than
	// sent as null — some servers reject an explicit null.
	if len(captured.Params) != 0 {
		t.Errorf("params = %s, want omitted", captured.Params)
	}
}

func TestCallAssignsIncreasingIDs(t *testing.T) {
	// The handler runs on the server's goroutine while the test reads on its
	// own, so the recorded ids need a lock rather than relying on the request
	// happening to complete first.
	var (
		mu  sync.Mutex
		ids []uint64
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("server could not decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		mu.Lock()
		ids = append(ids, req.ID)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"status":"healthy"}}`, req.ID)
	})

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := client.GetHealth(ctx); err != nil {
			t.Fatalf("GetHealth() call %d returned error: %v", i, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(ids) != 3 {
		t.Fatalf("recorded %d ids, want 3", len(ids))
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("id %d (%d) did not increase over previous (%d)", i, ids[i], ids[i-1])
		}
	}
}

func TestCallRPCError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"method not found"}}`)
	})

	_, err := client.GetHealth(context.Background())
	if err == nil {
		t.Fatal("GetHealth() returned nil error, want RPCError")
	}

	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("error %v (%T) is not a *RPCError", err, err)
	}
	if rpcErr.Code != -32601 {
		t.Errorf("Code = %d, want -32601", rpcErr.Code)
	}
	if rpcErr.Message != "method not found" {
		t.Errorf("Message = %q, want %q", rpcErr.Message, "method not found")
	}
}

func TestCallRPCErrorWithData(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid params","data":"key is not valid base64"}}`)
	})

	_, err := client.GetLedgerEntries(context.Background(), "not-base64")

	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("error %v (%T) is not a *RPCError", err, err)
	}
	if rpcErr.Data == "" {
		t.Error("Data is empty, want the server's data member")
	}
	// Data is raw JSON, so a string member arrives quoted.
	if want := `"key is not valid base64"`; rpcErr.Data != want {
		t.Errorf("Data = %s, want %s", rpcErr.Data, want)
	}
}

func TestCallHTTPError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, "rate limit exceeded")
	})

	_, err := client.GetHealth(context.Background())
	if err == nil {
		t.Fatal("GetHealth() returned nil error, want HTTPError")
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error %v (%T) is not a *HTTPError", err, err)
	}
	if httpErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want %d", httpErr.StatusCode, http.StatusTooManyRequests)
	}
	if httpErr.Body != "rate limit exceeded" {
		t.Errorf("Body = %q, want %q", httpErr.Body, "rate limit exceeded")
	}
}

func TestCallMalformedJSON(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{`)
	})

	if _, err := client.GetHealth(context.Background()); err == nil {
		t.Fatal("GetHealth() returned nil error, want a decode failure")
	}
}

func TestCallRejectsMismatchedID(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Deliberately not the id the client sent.
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":999999,"result":{"status":"healthy"}}`)
	})

	if _, err := client.GetHealth(context.Background()); err == nil {
		t.Fatal("GetHealth() accepted a response whose id did not match the request")
	}
}

func TestCallEmptyResult(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var captured capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d}`, captured.ID)
	})

	if _, err := client.GetHealth(context.Background()); err == nil {
		t.Fatal("GetHealth() accepted a response carrying neither result nor error")
	}
}

func TestCallHonoursContextCancellation(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler was reached despite a cancelled context")
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetHealth(ctx)
	if err == nil {
		t.Fatal("GetHealth() returned nil error, want context cancellation")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error %v does not wrap context.Canceled", err)
	}
}
