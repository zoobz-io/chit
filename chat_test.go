package chit_test

import (
	"context"
	"testing"

	"github.com/zoobz-io/chit"
	helpers "github.com/zoobz-io/chit/testing"
)

func TestNew(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)

	if chat.ID() == "" {
		t.Error("expected chat to have an ID")
	}

	if chat.History() == nil {
		t.Error("expected chat to have a history")
	}

	if chat.Config() == nil {
		t.Error("expected chat to have a config")
	}
}

func TestNewWithOptions(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	history := []chit.Message{{Role: "user", Content: "previous"}}

	chat := chit.New(processor, emitter,
		chit.WithHistory(history),
		chit.WithSystemPrompt("You are helpful."),
		chit.WithMetadata(map[string]any{"key": "value"}),
	)

	// History will have system prompt prepended, then the provided history
	h := chat.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 messages in history, got %d", len(h))
	}

	if chat.Config().SystemPrompt != "You are helpful." {
		t.Errorf("expected system prompt 'You are helpful.', got %q", chat.Config().SystemPrompt)
	}

	if chat.Config().Metadata["key"] != "value" {
		t.Error("expected metadata to be set")
	}
}

func TestHandle_Response(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)
	ctx := context.Background()

	err := chat.Handle(ctx, "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(emitter.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(emitter.Messages))
	}

	if emitter.Messages[0].Content != "Echo: Hello" {
		t.Errorf("expected 'Echo: Hello', got %q", emitter.Messages[0].Content)
	}

	if emitter.Messages[0].Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", emitter.Messages[0].Role)
	}
}

func TestHandle_Yield(t *testing.T) {
	processor := helpers.YieldingProcessor("What is your name?", "Hello")
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)
	ctx := context.Background()

	// First call should yield
	err := chat.Handle(ctx, "Start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !chat.HasContinuation() {
		t.Error("expected chat to have continuation after yield")
	}

	if len(emitter.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(emitter.Messages))
	}

	if emitter.Messages[0].Content != "What is your name?" {
		t.Errorf("expected 'What is your name?', got %q", emitter.Messages[0].Content)
	}

	// Second call should resume and complete
	err = chat.Handle(ctx, "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chat.HasContinuation() {
		t.Error("expected chat to not have continuation after response")
	}

	if len(emitter.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(emitter.Messages))
	}

	if emitter.Messages[1].Content != "Hello: Alice" {
		t.Errorf("expected 'Hello: Alice', got %q", emitter.Messages[1].Content)
	}
}

func TestHandle_ProcessorReceivesHistory(t *testing.T) {
	var receivedHistory []chit.Message

	processor := &helpers.MockProcessor{
		ProcessFunc: func(_ context.Context, _ string, history []chit.Message) (chit.Result, error) {
			receivedHistory = history
			return &chit.Response{Content: "ok"}, nil
		},
	}
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)
	ctx := context.Background()

	err := chat.Handle(ctx, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedHistory == nil {
		t.Error("expected processor to receive history")
	}

	// History should contain the user message that was just added
	if len(receivedHistory) != 1 {
		t.Fatalf("expected 1 message in history, got %d", len(receivedHistory))
	}

	if receivedHistory[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", receivedHistory[0].Role)
	}

	if receivedHistory[0].Content != "test" {
		t.Errorf("expected content 'test', got %q", receivedHistory[0].Content)
	}
}

func TestHandle_HistoryIsReadOnly(t *testing.T) {
	processor := &helpers.MockProcessor{
		ProcessFunc: func(_ context.Context, _ string, history []chit.Message) (chit.Result, error) {
			// Attempt to modify history - should not affect the chat's history
			if len(history) > 0 {
				history[0].Content = "modified"
			}
			return &chit.Response{Content: "ok"}, nil
		},
	}
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)
	ctx := context.Background()

	_ = chat.Handle(ctx, "original")

	// History should be unaffected by processor's modification attempt
	messages := chat.History()
	if messages[0].Content != "original" {
		t.Errorf("expected history to be unmodified, got %q", messages[0].Content)
	}
}

func TestHandle_AddsMessagesToHistory(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)
	ctx := context.Background()

	err := chat.Handle(ctx, "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	messages := chat.History()

	// Should have user message and assistant response
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages in history, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("expected first message role 'user', got %q", messages[0].Role)
	}

	if messages[0].Content != "Hello" {
		t.Errorf("expected first message content 'Hello', got %q", messages[0].Content)
	}

	if messages[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got %q", messages[1].Role)
	}
}
