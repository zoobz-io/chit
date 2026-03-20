package chit_test

import (
	"testing"

	"github.com/zoobz-io/chit"
)

func TestChatRequest_Fields(t *testing.T) {
	history := []chit.Message{{Role: "user", Content: "hello"}}
	req := &chit.ChatRequest{
		Input:     "hello",
		History:   history,
		ChatID:    "chat-123",
		RequestID: "req-456",
	}

	if req.Input != "hello" {
		t.Errorf("expected input 'hello', got %q", req.Input)
	}

	if len(req.History) != 1 {
		t.Error("expected history to be set")
	}

	if req.ChatID != "chat-123" {
		t.Errorf("expected chatID 'chat-123', got %q", req.ChatID)
	}

	if req.RequestID != "req-456" {
		t.Errorf("expected requestID 'req-456', got %q", req.RequestID)
	}
}

func TestChatRequest_Clone(t *testing.T) {
	history := []chit.Message{{Role: "user", Content: "hello"}}
	original := &chit.ChatRequest{
		Input:     "original",
		History:   history,
		ChatID:    "chat-123",
		RequestID: "req-456",
		Result:    &chit.Response{Content: "response"},
	}

	cloned := original.Clone()

	// Verify fields are copied
	if cloned.Input != original.Input {
		t.Errorf("expected cloned input %q, got %q", original.Input, cloned.Input)
	}

	if cloned.ChatID != original.ChatID {
		t.Errorf("expected cloned chatID %q, got %q", original.ChatID, cloned.ChatID)
	}

	if cloned.RequestID != original.RequestID {
		t.Errorf("expected cloned requestID %q, got %q", original.RequestID, cloned.RequestID)
	}

	// History is copied (not shared)
	if len(cloned.History) != len(original.History) {
		t.Error("expected history to be copied")
	}

	// Modifying cloned history doesn't affect original
	cloned.History[0].Content = "modified"
	if original.History[0].Content == "modified" {
		t.Error("modifying cloned history should not affect original")
	}

	// Modifying clone input doesn't affect original
	cloned.Input = "modified"
	if original.Input == "modified" {
		t.Error("modifying clone should not affect original")
	}
}

func TestChatRequest_ResultField(t *testing.T) {
	req := &chit.ChatRequest{
		Input: "test",
	}

	// Initially nil
	if req.Result != nil {
		t.Error("expected Result to be nil initially")
	}

	// Can set Response
	req.Result = &chit.Response{Content: "hello"}
	resp, ok := req.Result.(*chit.Response)
	if !ok {
		t.Fatal("expected Result to be *Response")
	}
	if resp.Content != "hello" {
		t.Errorf("expected content 'hello', got %q", resp.Content)
	}

	// Can set Yield
	req.Result = &chit.Yield{Prompt: "question?"}
	yield, ok := req.Result.(*chit.Yield)
	if !ok {
		t.Fatal("expected Result to be *Yield")
	}
	if yield.Prompt != "question?" {
		t.Errorf("expected prompt 'question?', got %q", yield.Prompt)
	}
}
