package chit

import "context"

// contextKey is used to store values in context without collisions.
type contextKey string

const (
	emitterKey contextKey = "chit.emitter"
)

// WithEmitter returns a new context with the given Emitter attached.
// Processors can retrieve the Emitter via EmitterFromContext to push
// resources or stream partial responses during processing.
func WithEmitter(ctx context.Context, e Emitter) context.Context {
	return context.WithValue(ctx, emitterKey, e)
}

// EmitterFromContext retrieves the Emitter from the context.
// Returns nil if no Emitter is present.
func EmitterFromContext(ctx context.Context) Emitter {
	if e, ok := ctx.Value(emitterKey).(Emitter); ok {
		return e
	}
	return nil
}
