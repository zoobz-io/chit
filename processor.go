// Package chit provides conversation lifecycle management for LLM-powered chat applications.
//
// Chit is a focused conversation controller that orchestrates pluggable processors
// for reasoning, action execution, and data retrieval. It handles turn-taking,
// session management, and response streaming.
package chit

import "context"

// Processor handles reasoning and execution logic for a conversation.
// It receives user input and conversation history (read-only) for context,
// and returns a Result. How the processor communicates with LLMs is an
// internal implementation detail.
type Processor interface {
	// Process handles user input and returns a result.
	// History provides the user conversation for context (read-only).
	// Returns a Response when complete, or a Yield when awaiting input.
	Process(ctx context.Context, input string, history []Message) (Result, error)
}

// ProcessorFunc is an adapter that allows using a function as a Processor.
type ProcessorFunc func(ctx context.Context, input string, history []Message) (Result, error)

// Process implements the Processor interface.
func (f ProcessorFunc) Process(ctx context.Context, input string, history []Message) (Result, error) {
	return f(ctx, input, history)
}
