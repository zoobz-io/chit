package chit

import (
	"context"

	"github.com/zoobzio/pipz"
)

var middlewareCallbackID = pipz.NewIdentity("chit:middleware", "User middleware callback")

// NewMiddleware creates middleware from a callback function.
// The callback can inspect/modify the request or return an error to halt processing.
// For advanced use cases, create pipz processors directly and pass to WithMiddleware.
func NewMiddleware(fn func(context.Context, *ChatRequest) (*ChatRequest, error)) pipz.Chainable[*ChatRequest] {
	return pipz.Apply(middlewareCallbackID, fn)
}
