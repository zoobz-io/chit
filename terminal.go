package chit

import (
	"context"

	"github.com/zoobz-io/pipz"
)

// Identities for terminal processors.
var (
	terminalID             = pipz.NewIdentity("chit:terminal", "Processor terminal")
	continuationTerminalID = pipz.NewIdentity("chit:continuation-terminal", "Continuation terminal")
)

// NewTerminal creates the terminal processor that calls the Processor.
// This is the innermost layer of the pipeline that performs the actual processing.
func NewTerminal(processor Processor) pipz.Chainable[*ChatRequest] {
	return pipz.Apply(terminalID, func(ctx context.Context, req *ChatRequest) (*ChatRequest, error) {
		result, err := processor.Process(ctx, req.Input, req.History)
		if err != nil {
			return req, err
		}
		req.Result = result
		return req, nil
	})
}

// NewContinuationTerminal creates a terminal that calls the given continuation.
// Used to wrap continuation calls with the same reliability options as fresh calls.
func NewContinuationTerminal(cont Continuation) pipz.Chainable[*ChatRequest] {
	return pipz.Apply(continuationTerminalID, func(ctx context.Context, req *ChatRequest) (*ChatRequest, error) {
		result, err := cont(ctx, req.Input, req.History)
		if err != nil {
			return req, err
		}
		req.Result = result
		return req, nil
	})
}
