package soroban

import "context"

// Event represents a Soroban contract event.
type Event struct {
	// ContractAddress is the contract address that emitted the event.
	ContractAddress string `json:"contractAddress"`

	// Topics are the event topics, typically including the event name and
	// indexed parameters.
	Topics []string `json:"topics"`

	// Data is the event data, base64-encoded XDR.
	Data string `json:"data"`

	// Ledger is the ledger number where the event was observed.
	Ledger uint32 `json:"ledger"`

	// LedgerCloseTime is the UNIX timestamp of when the ledger closed.
	LedgerCloseTime uint64 `json:"ledgerCloseTime"`

	// ID is a unique identifier for the event.
	ID string `json:"id"`

	// PagingToken can be used for pagination in subsequent requests.
	PagingToken string `json:"pagingToken,omitempty"`
}

// GetEventsRequest defines the parameters for the getEvents RPC call.
type GetEventsRequest struct {
	// Cursor specifies where to start fetching events from.
	Cursor string `json:"cursor,omitempty"`

	// Limit specifies the maximum number of events to return.
	Limit int32 `json:"limit,omitempty"`

	// Order specifies the chronological order of events: "asc" or "desc".
	Order string `json:"order,omitempty"`

	// Filters specifies criteria to filter events.
	Filters *EventFilters `json:"filters,omitempty"`
}

// EventFilters defines filtering options for getEvents.
type EventFilters struct {
	// ContractIDs is an array of contract addresses to filter by.
	ContractIDs []string `json:"contractIds,omitempty"`

	// Types is an array of event types to filter by.
	Types []string `json:"types,omitempty"`

	// StartLedger is the minimum ledger (inclusive) to consider.
	StartLedger uint32 `json:"startLedger,omitempty"`

	// StopLedger is the maximum ledger (inclusive) to consider.
	StopLedger uint32 `json:"stopLedger,omitempty"`
	// TransactionID is the transaction hash to filter events by.
	TransactionID string `json:"txId,omitempty"`

	// LedgerBounds specifies a ledger range to consider.
	LedgerBounds *LedgerBounds `json:"ledgerBounds,omitempty"`
}

// LedgerBounds specifies a range of ledgers.
type LedgerBounds struct {
	// MinLedger is the minimum ledger (inclusive).
	MinLedger uint32 `json:"minLedger"`

	// MaxLedger is the maximum ledger (inclusive).
	MaxLedger uint32 `json:"maxLedger"`
}

// GetEventsResponse is the result of getEvents.
type GetEventsResponse struct {
	// LatestLedger is the ledger the query was served against.
	LatestLedger uint32 `json:"latestLedger"`

	// Events is the array of events that match the request.
	Events []Event `json:"events"`

	// Cursor is the pagination cursor for fetching more results.
	Cursor string `json:"cursor,omitempty"`
}

// GetEvents calls getEvents, returning contract events that match the given
// filters and pagination options.
//
// The request parameter can be nil to use default values (no filters, no
// pagination, server-defined limit).
func (c *Client) GetEvents(ctx context.Context, req *GetEventsRequest) (*GetEventsResponse, error) {
	var out GetEventsResponse
	if err := c.call(ctx, "getEvents", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
