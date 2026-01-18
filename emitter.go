package chit

import "context"

// Emitter handles all output to the client.
// It supports both streaming response content and pushing structured resources.
type Emitter interface {
	// Emit streams response content (conversational text) to the client.
	Emit(ctx context.Context, msg Message) error

	// Push sends structured resources synchronously to the client.
	// Used for delivering fetched data, context, or tool results.
	Push(ctx context.Context, resource Resource) error

	// Close signals the end of output and releases resources.
	Close() error
}

// Message is a streamed response unit in a conversation.
type Message struct {
	// Role identifies the message sender (e.g., "assistant", "system").
	Role string

	// Content is the text content of the message.
	Content string

	// Metadata contains optional additional information about the message.
	Metadata map[string]any
}

// Resource is structured data pushed to the client.
// Used for delivering fetched data, context enrichment, or tool results.
type Resource struct {
	// Type identifies the kind of resource (e.g., "data", "context", "tool_result").
	Type string

	// URI is the scio URI if this resource was retrieved from a data source.
	URI string

	// Payload is the structured data being delivered.
	Payload any

	// Metadata contains optional additional information about the resource.
	Metadata map[string]any
}
