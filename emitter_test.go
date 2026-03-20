package chit_test

import (
	"testing"

	"github.com/zoobz-io/chit"
)

func TestMessage_Fields(t *testing.T) {
	msg := chit.Message{
		Role:     "assistant",
		Content:  "Hello, world!",
		Metadata: map[string]any{"key": "value"},
	}

	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", msg.Role)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got %q", msg.Content)
	}

	if msg.Metadata["key"] != "value" {
		t.Error("expected metadata to contain key=value")
	}
}

func TestMessage_EmptyMetadata(t *testing.T) {
	msg := chit.Message{
		Role:    "user",
		Content: "test",
	}

	// Nil metadata should not panic on access
	if msg.Metadata != nil {
		t.Error("expected nil metadata")
	}
}

func TestResource_Fields(t *testing.T) {
	resource := chit.Resource{
		Type:     "data",
		URI:      "db://users/123",
		Payload:  map[string]string{"name": "Alice"},
		Metadata: map[string]any{"fetched_at": "2024-01-01"},
	}

	if resource.Type != "data" {
		t.Errorf("expected type 'data', got %q", resource.Type)
	}

	if resource.URI != "db://users/123" {
		t.Errorf("expected URI 'db://users/123', got %q", resource.URI)
	}

	payload, ok := resource.Payload.(map[string]string)
	if !ok {
		t.Fatal("expected payload to be map[string]string")
	}

	if payload["name"] != "Alice" {
		t.Errorf("expected payload name 'Alice', got %q", payload["name"])
	}

	if resource.Metadata["fetched_at"] != "2024-01-01" {
		t.Error("expected metadata to contain fetched_at")
	}
}

func TestResource_EmptyFields(t *testing.T) {
	resource := chit.Resource{
		Type: "context",
	}

	if resource.URI != "" {
		t.Error("expected empty URI")
	}

	if resource.Payload != nil {
		t.Error("expected nil payload")
	}

	if resource.Metadata != nil {
		t.Error("expected nil metadata")
	}
}

func TestResource_AnyPayload(t *testing.T) {
	// Resource.Payload is `any`, so it can hold various types

	t.Run("string", func(t *testing.T) {
		resource := chit.Resource{Type: "test", Payload: "simple string"}
		if resource.Payload != "simple string" {
			t.Error("expected string payload")
		}
	})

	t.Run("int", func(t *testing.T) {
		resource := chit.Resource{Type: "test", Payload: 42}
		if resource.Payload != 42 {
			t.Error("expected int payload")
		}
	})

	t.Run("slice", func(t *testing.T) {
		payload := []string{"a", "b", "c"}
		resource := chit.Resource{Type: "test", Payload: payload}
		slice, ok := resource.Payload.([]string)
		if !ok {
			t.Fatal("expected slice payload")
		}
		if len(slice) != 3 || slice[0] != "a" {
			t.Error("slice payload mismatch")
		}
	})

	t.Run("struct", func(t *testing.T) {
		type testStruct struct{ Name string }
		payload := testStruct{"test"}
		resource := chit.Resource{Type: "test", Payload: payload}
		s, ok := resource.Payload.(testStruct)
		if !ok {
			t.Fatal("expected struct payload")
		}
		if s.Name != "test" {
			t.Error("struct payload mismatch")
		}
	})

	t.Run("nil", func(t *testing.T) {
		resource := chit.Resource{Type: "test", Payload: nil}
		if resource.Payload != nil {
			t.Error("expected nil payload")
		}
	})
}
