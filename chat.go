package chit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
)

// Chat is a conversation lifecycle controller.
// It orchestrates a pluggable Processor for reasoning and execution,
// manages turn-taking via continuations, and streams responses through an Emitter.
type Chat struct {
	id           string
	history      []Message
	config       *Config
	processor    Processor
	emitter      Emitter
	continuation Continuation
	pipeline     pipz.Chainable[*ChatRequest]
	pipelineOpts []PipelineOption
	mu           sync.Mutex
}

// Config holds configuration for a Chat.
type Config struct {
	// SystemPrompt is the system message that guides conversation behavior.
	SystemPrompt string

	// Metadata contains optional additional configuration data.
	Metadata map[string]any
}

// New creates a new Chat with the given processor and emitter.
// Both processor and emitter are required.
// Use functional options to configure the Chat further.
func New(processor Processor, emitter Emitter, opts ...Option) *Chat {
	c := &Chat{
		id:        uuid.New().String(),
		processor: processor,
		emitter:   emitter,
		config:    &Config{},
		history:   make([]Message, 0),
	}

	// Apply options (both config and pipeline options)
	for _, opt := range opts {
		opt(c)
	}

	// Apply system prompt to history if configured
	if c.config.SystemPrompt != "" {
		c.history = append(c.history, Message{
			Role:    "system",
			Content: c.config.SystemPrompt,
		})
	}

	// Build pipeline: terminal first, then wrap with options
	pipeline := NewTerminal(processor)
	for _, opt := range c.pipelineOpts {
		pipeline = opt(pipeline)
	}
	c.pipeline = pipeline

	// Emit chat created signal
	capitan.Emit(context.Background(), ChatCreated,
		FieldChatID.Field(c.id),
	)

	return c
}

// ID returns the unique identifier for this Chat.
func (c *Chat) ID() string {
	return c.id
}

// History returns a copy of the conversation history.
func (c *Chat) History() []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	history := make([]Message, len(c.history))
	copy(history, c.history)
	return history
}

// Config returns the configuration for this Chat.
func (c *Chat) Config() *Config {
	return c.config
}

// HasContinuation returns true if the Chat is awaiting input
// to resume a previously yielded processing step.
func (c *Chat) HasContinuation() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.continuation != nil
}

// GetPipeline returns the internal pipeline for composition.
// This implements ChatProvider for use with WithFallback.
func (c *Chat) GetPipeline() pipz.Chainable[*ChatRequest] {
	return c.pipeline
}

// buildContinuationPipeline creates a pipeline for continuation calls.
// It wraps the continuation with the same reliability options as the main pipeline,
// ensuring retry, timeout, rate limiting, and middleware apply to resumed processing.
func (c *Chat) buildContinuationPipeline(cont Continuation) pipz.Chainable[*ChatRequest] {
	pipeline := NewContinuationTerminal(cont)
	for _, opt := range c.pipelineOpts {
		pipeline = opt(pipeline)
	}
	return pipeline
}

// Handle processes user input through the conversation lifecycle.
//
// The lifecycle is:
//  1. Add user input to history
//  2. If a continuation exists, resume with the input
//  3. Otherwise, process the input through the pipeline
//  4. Add response to history and emit via Emitter
//  5. If result is a Yield, store the continuation for next input
func (c *Chat) Handle(ctx context.Context, input string) error {
	c.mu.Lock()

	// Emit input received signal
	capitan.Emit(ctx, InputReceived,
		FieldChatID.Field(c.id),
		FieldInput.Field(input),
		FieldInputSize.Field(len(input)),
	)

	// Inject emitter into context for processor access
	ctx = WithEmitter(ctx, c.emitter)

	// Add user message to history
	c.history = append(c.history, Message{
		Role:    "user",
		Content: input,
	})

	// Snapshot history for read-only access
	history := make([]Message, len(c.history))
	copy(history, c.history)

	var result Result
	var err error

	isResume := c.continuation != nil
	start := time.Now()

	// Emit processing started signal
	capitan.Emit(ctx, ProcessingStarted,
		FieldChatID.Field(c.id),
	)

	// Build request for pipeline processing
	req := &ChatRequest{
		Input:     input,
		History:   history,
		ChatID:    c.id,
		RequestID: uuid.New().String(),
	}

	if isResume {
		// Emit turn resumed signal
		capitan.Emit(ctx, TurnResumed,
			FieldChatID.Field(c.id),
		)

		// Build continuation pipeline with same reliability options
		cont := c.continuation
		c.continuation = nil
		c.mu.Unlock()

		contPipeline := c.buildContinuationPipeline(cont)
		processed, pipelineErr := contPipeline.Process(ctx, req)
		if pipelineErr != nil {
			err = pipelineErr
		} else {
			result = processed.Result
		}
	} else {
		c.mu.Unlock()

		// Process through main pipeline
		processed, pipelineErr := c.pipeline.Process(ctx, req)
		if pipelineErr != nil {
			err = pipelineErr
		} else {
			result = processed.Result
		}
	}

	duration := time.Since(start)

	if err != nil {
		// Emit processing failed signal
		capitan.Emit(ctx, ProcessingFailed,
			FieldChatID.Field(c.id),
			FieldProcessingDuration.Field(duration),
			FieldError.Field(err),
		)
		return err
	}

	// Emit processing completed signal
	capitan.Emit(ctx, ProcessingCompleted,
		FieldChatID.Field(c.id),
		FieldProcessingDuration.Field(duration),
	)

	return c.handleResult(ctx, result)
}

// handleResult processes the result of processing and takes appropriate action.
func (c *Chat) handleResult(ctx context.Context, result Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch r := result.(type) {
	case *Response:
		// Add assistant response to history
		c.history = append(c.history, Message{
			Role:    "assistant",
			Content: r.Content,
		})

		// Emit response signal
		capitan.Emit(ctx, ResponseEmitted,
			FieldChatID.Field(c.id),
			FieldRole.Field("assistant"),
			FieldContentSize.Field(len(r.Content)),
		)

		// Emit the response via emitter
		return c.emitter.Emit(ctx, Message{
			Role:     "assistant",
			Content:  r.Content,
			Metadata: r.Metadata,
		})

	case *Yield:
		// Add yield prompt to history as assistant message
		c.history = append(c.history, Message{
			Role:    "assistant",
			Content: r.Prompt,
		})

		// Store continuation for next input
		c.continuation = r.Continuation

		// Emit turn yielded signal
		capitan.Emit(ctx, TurnYielded,
			FieldChatID.Field(c.id),
			FieldPrompt.Field(r.Prompt),
		)

		// Emit the yield prompt via emitter
		return c.emitter.Emit(ctx, Message{
			Role:     "assistant",
			Content:  r.Prompt,
			Metadata: r.Metadata,
		})

	default:
		return ErrUnknownResultType
	}
}
