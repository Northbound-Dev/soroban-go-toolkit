package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Northbound-Dev/soroban-go-toolkit/pkg/soroban"
)

// eventsOutput is the CLI's own JSON shape for events read.
//
// The client returns types that would serialise into something verbose
// and tied to the SDK's internal representation. Declaring the output
// explicitly keeps --json a stable contract for scripts.
type eventsOutput struct {
	LatestLedger uint32            `json:"latestLedger"`
	Events       []eventOutput     `json:"events"`
	Cursor       string            `json:"cursor,omitempty"`
}

// eventOutput is the CLI's own JSON shape for a single event.
type eventOutput struct {
	ContractAddress string   `json:"contractAddress"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	Ledger          uint32   `json:"ledger"`
	LedgerCloseTime uint64   `json:"ledgerCloseTime"`
	ID              string   `json:"id"`
	PagingToken     string   `json:"pagingToken,omitempty"`
}

func newEventsOutput(resp *soroban.GetEventsResponse) eventsOutput {
	eventOutputs := make([]eventOutput, len(resp.Events))
	for i, event := range resp.Events {
		eventOutputs[i] = eventOutput{
			ContractAddress: event.ContractAddress,
			Topics:          event.Topics,
			Data:            event.Data,
			Ledger:          event.Ledger,
			LedgerCloseTime: event.LedgerCloseTime,
			ID:              event.ID,
			PagingToken:     event.PagingToken,
		}
	}

	return eventsOutput{
		LatestLedger: resp.LatestLedger,
		Events:       eventOutputs,
		Cursor:       resp.Cursor,
	}
}

func printEvents(w io.Writer, out eventsOutput) {
	fmt.Fprintf(w, "latest ledger: %d\n", out.LatestLedger)
	fmt.Fprintf(w, "found:         %d events\n", len(out.Events))
	if out.Cursor != "" {
		fmt.Fprintf(w, "cursor:        %s\n", out.Cursor)
	}
	fmt.Fprintln(w)

	for i, event := range out.Events {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "event %d:\n", i+1)
		fmt.Fprintf(w, "  contract:    %s\n", event.ContractAddress)
		fmt.Fprintf(w, "  topics:      %q\n", event.Topics)
		fmt.Fprintf(w, "  data:        %s\n", event.Data)
		fmt.Fprintf(w, "  ledger:      %d\n", event.Ledger)
		fmt.Fprintf(w, "  ledger time: %d\n", event.LedgerCloseTime)
		fmt.Fprintf(w, "  id:          %s\n", event.ID)
		if event.PagingToken != "" {
			fmt.Fprintf(w, "  paging tok:  %s\n", event.PagingToken)
		}
	}
}

// parseEventOrder maps the flag value for event order.
func parseEventOrder(value string) (string, error) {
	switch strings.ToLower(value) {
	case "asc", "desc":
		return strings.ToLower(value), nil
	default:
		return "", fmt.Errorf("invalid order %q, must be asc or desc", value)
	}
}

// newEventsCommand creates the sorobango events command.
func newEventsCommand(opts *options) *cobra.Command {
	var (
		cursor      string
		limit       int32
		order       string
		contractIDs stringsliceValue
		eventTypes  stringsliceValue
		startLedger uint32
		stopLedger  uint32
	)

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Query Soroban contract events",
		Long: `Query contract events emitted on the Stellar Soroban network.

Events can be filtered by contract address, event type, ledger range, and
pagination options. This is useful for building indexers, monitoring dApp
activity, or analyzing contract behavior over time.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.client()
			if err != nil {
				return err
			}

			var filters *soroban.EventFilters
			if len(contractIDs) > 0 || len(eventTypes) > 0 || startLedger > 0 || stopLedger > 0 {
				filters = &soroban.EventFilters{
					ContractIDs: contractIDs,
					Types:       eventTypes,
					StartLedger: startLedger,
					StopLedger:  stopLedger,
				}
				// If both StartLedger/StopLedger and LedgerBounds are set,
				// StartLedger/StopLedger takes precedence (more specific)
			}

			req := &soroban.GetEventsRequest{
				Cursor:  cursor,
				Limit:   limit,
				Order:   order,
				Filters: filters,
			}

			resp, err := client.GetEvents(cmd.Context(), req)
			if err != nil {
				return err
			}

			out := newEventsOutput(resp)
			return emit(cmd, opts.asJSON, out, func(w io.Writer) {
				printEvents(w, out)
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&cursor, "cursor", "", "pagination cursor (from previous response)")
	flags.Int32Var(&limit, "limit", 0, "maximum number of events to return (0 = server default)")
	flags.StringVar(&order, "order", "", "event order: asc or desc (empty = server default)")
	flags.Var(&contractIDs, "contract-id", "contract address to filter by (can be repeated)")
	flags.Var(&eventTypes, "type", "event type to filter by (can be repeated)")
	flags.Uint32Var(&startLedger, "start-ledger", 0, "minimum ledger (inclusive) to consider")
	flags.Uint32Var(&stopLedger, "stop-ledger", 0, "maximum ledger (inclusive) to consider")

	return cmd
}

// stringsliceValue is a cobra.Value that accumulates string values into a slice.
type stringsliceValue []string

func (s *stringsliceValue) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *stringsliceValue) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringsliceValue) Type() string {
	return "string"
}