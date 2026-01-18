package chit_test

import (
	"testing"

	"github.com/zoobzio/chit"
	helpers "github.com/zoobzio/chit/testing"
)

func TestWithConfig(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	config := &chit.Config{
		SystemPrompt: "You are a test assistant.",
		Metadata:     map[string]any{"version": "1.0"},
	}

	chat := chit.New(processor, emitter, chit.WithConfig(config))

	if chat.Config() != config {
		t.Error("expected WithConfig to set the config")
	}

	if chat.Config().SystemPrompt != "You are a test assistant." {
		t.Errorf("expected system prompt, got %q", chat.Config().SystemPrompt)
	}

	if chat.Config().Metadata["version"] != "1.0" {
		t.Error("expected metadata to be set")
	}
}

func TestWithSystemPrompt(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter, chit.WithSystemPrompt("Be helpful."))

	if chat.Config().SystemPrompt != "Be helpful." {
		t.Errorf("expected system prompt 'Be helpful.', got %q", chat.Config().SystemPrompt)
	}
}

func TestWithSystemPrompt_AddsToHistory(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter, chit.WithSystemPrompt("System message."))

	messages := chat.History()
	if len(messages) != 1 {
		t.Fatalf("expected 1 message in history, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("expected role 'system', got %q", messages[0].Role)
	}

	if messages[0].Content != "System message." {
		t.Errorf("expected content 'System message.', got %q", messages[0].Content)
	}
}

func TestWithHistory(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	history := []chit.Message{{Role: "user", Content: "pre-existing message"}}

	chat := chit.New(processor, emitter, chit.WithHistory(history))

	messages := chat.History()
	if len(messages) != 1 {
		t.Fatalf("expected 1 pre-existing message, got %d", len(messages))
	}

	if messages[0].Content != "pre-existing message" {
		t.Errorf("expected pre-existing message, got %q", messages[0].Content)
	}
}

func TestWithMetadata(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	metadata := map[string]any{
		"user_id":    "123",
		"session_id": "abc",
	}

	chat := chit.New(processor, emitter, chit.WithMetadata(metadata))

	if chat.Config().Metadata["user_id"] != "123" {
		t.Error("expected user_id in metadata")
	}

	if chat.Config().Metadata["session_id"] != "abc" {
		t.Error("expected session_id in metadata")
	}
}

func TestMultipleOptions(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	history := []chit.Message{{Role: "user", Content: "previous"}}

	chat := chit.New(processor, emitter,
		chit.WithHistory(history),
		chit.WithSystemPrompt("Be concise."),
		chit.WithMetadata(map[string]any{"key": "value"}),
	)

	// History will have system prompt prepended
	h := chat.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(h))
	}

	if chat.Config().SystemPrompt != "Be concise." {
		t.Error("expected system prompt to be set")
	}

	if chat.Config().Metadata["key"] != "value" {
		t.Error("expected metadata to be set")
	}
}

func TestWithMetadata_NilConfig(t *testing.T) {
	// WithMetadata should create config if nil
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter, chit.WithMetadata(map[string]any{"test": true}))

	if chat.Config() == nil {
		t.Fatal("expected config to be created")
	}

	if chat.Config().Metadata["test"] != true {
		t.Error("expected metadata to be set")
	}
}
