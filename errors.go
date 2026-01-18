package chit

import "errors"

// Sentinel errors for chit operations.
var (
	// ErrUnknownResultType is returned when a Processor returns a Result
	// that is neither a Response nor a Yield.
	ErrUnknownResultType = errors.New("chit: unknown result type")

	// ErrNilProcessor is returned when attempting to create a Chat without a Processor.
	ErrNilProcessor = errors.New("chit: processor is required")

	// ErrNilEmitter is returned when attempting to create a Chat without an Emitter.
	ErrNilEmitter = errors.New("chit: emitter is required")

	// ErrEmitterClosed is returned when attempting to emit after the Emitter is closed.
	ErrEmitterClosed = errors.New("chit: emitter is closed")
)
