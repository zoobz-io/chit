package chit

import "context"

// Result represents the outcome of processing user input.
// A result is either complete (Response) or yielded (Yield).
type Result interface {
	// IsComplete returns true if processing finished with a response to emit.
	IsComplete() bool

	// IsYielded returns true if processing paused awaiting further input.
	IsYielded() bool
}

// Response is a complete result with content to emit to the user.
type Response struct {
	// Content is the response text to emit.
	Content string

	// Metadata contains optional additional information about the response.
	Metadata map[string]any
}

// IsComplete returns true for Response.
func (r *Response) IsComplete() bool { return true }

// IsYielded returns false for Response.
func (r *Response) IsYielded() bool { return false }

// Yield indicates processing has paused and is awaiting further input.
// The Continuation function resumes processing when input arrives.
type Yield struct {
	// Prompt is the message to emit while awaiting input (e.g., a question).
	Prompt string

	// Continuation is the function to call with the next input to resume processing.
	Continuation Continuation

	// Metadata contains optional additional information about the yield.
	Metadata map[string]any
}

// IsComplete returns false for Yield.
func (y *Yield) IsComplete() bool { return false }

// IsYielded returns true for Yield.
func (y *Yield) IsYielded() bool { return true }

// Continuation is a function that resumes processing with new input.
// History provides the updated user conversation since the yield.
type Continuation func(ctx context.Context, input string, history []Message) (Result, error)
