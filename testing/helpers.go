// Package testing provides test helpers for chit.
package testing

import (
	"context"
	"sync"

	"github.com/zoobz-io/chit"
)

// MockProcessor is a test implementation of chit.Processor.
type MockProcessor struct {
	// ProcessFunc is called when Process is invoked.
	// Set this to control the mock's behavior.
	ProcessFunc func(ctx context.Context, input string, history []chit.Message) (chit.Result, error)

	// Calls records all calls to Process for verification.
	Calls []ProcessCall

	mu sync.Mutex
}

// ProcessCall records a single call to Process.
type ProcessCall struct {
	Input   string
	History []chit.Message
}

// Process implements chit.Processor.
func (m *MockProcessor) Process(ctx context.Context, input string, history []chit.Message) (chit.Result, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, ProcessCall{Input: input, History: history})
	m.mu.Unlock()

	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, input, history)
	}

	// Default: return empty response
	return &chit.Response{Content: ""}, nil
}

// MockEmitter is a test implementation of chit.Emitter.
type MockEmitter struct {
	// EmitFunc is called when Emit is invoked.
	EmitFunc func(ctx context.Context, msg chit.Message) error

	// PushFunc is called when Push is invoked.
	PushFunc func(ctx context.Context, resource chit.Resource) error

	// CloseFunc is called when Close is invoked.
	CloseFunc func() error

	// Messages records all emitted messages.
	Messages []chit.Message

	// Resources records all pushed resources.
	Resources []chit.Resource

	// Closed indicates whether Close was called.
	Closed bool

	mu sync.Mutex
}

// Emit implements chit.Emitter.
func (m *MockEmitter) Emit(ctx context.Context, msg chit.Message) error {
	m.mu.Lock()
	m.Messages = append(m.Messages, msg)
	m.mu.Unlock()

	if m.EmitFunc != nil {
		return m.EmitFunc(ctx, msg)
	}
	return nil
}

// Push implements chit.Emitter.
func (m *MockEmitter) Push(ctx context.Context, resource chit.Resource) error {
	m.mu.Lock()
	m.Resources = append(m.Resources, resource)
	m.mu.Unlock()

	if m.PushFunc != nil {
		return m.PushFunc(ctx, resource)
	}
	return nil
}

// Close implements chit.Emitter.
func (m *MockEmitter) Close() error {
	m.mu.Lock()
	m.Closed = true
	m.mu.Unlock()

	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// CollectingEmitter collects all emitted content for easy assertion.
type CollectingEmitter struct {
	Messages  []chit.Message
	Resources []chit.Resource
	mu        sync.Mutex
}

// Emit implements chit.Emitter.
func (e *CollectingEmitter) Emit(_ context.Context, msg chit.Message) error {
	e.mu.Lock()
	e.Messages = append(e.Messages, msg)
	e.mu.Unlock()
	return nil
}

// Push implements chit.Emitter.
func (e *CollectingEmitter) Push(_ context.Context, resource chit.Resource) error {
	e.mu.Lock()
	e.Resources = append(e.Resources, resource)
	e.mu.Unlock()
	return nil
}

// Close implements chit.Emitter.
func (e *CollectingEmitter) Close() error {
	return nil
}

// Content returns all emitted message content concatenated.
func (e *CollectingEmitter) Content() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	var content string
	for _, msg := range e.Messages {
		content += msg.Content
	}
	return content
}

// EchoProcessor returns a processor that echoes the input as a response.
func EchoProcessor() chit.Processor {
	return chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Response{Content: "Echo: " + input}, nil
	})
}

// YieldingProcessor returns a processor that yields on the first call
// and responds on the second call (via continuation).
func YieldingProcessor(prompt string, finalResponse string) chit.Processor {
	return chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: prompt,
			Continuation: func(_ context.Context, resumeInput string, _ []chit.Message) (chit.Result, error) {
				return &chit.Response{Content: finalResponse + ": " + resumeInput}, nil
			},
		}, nil
	})
}
