package chit_test

import (
	"context"
	"testing"

	"github.com/zoobz-io/chit"
)

func TestResponse_IsComplete(t *testing.T) {
	r := &chit.Response{Content: "test"}

	if !r.IsComplete() {
		t.Error("expected Response.IsComplete() to return true")
	}

	if r.IsYielded() {
		t.Error("expected Response.IsYielded() to return false")
	}
}

func TestYield_IsYielded(t *testing.T) {
	y := &chit.Yield{
		Prompt: "test",
		Continuation: func(_ context.Context, _ string, _ []chit.Message) (chit.Result, error) {
			return &chit.Response{Content: "continued"}, nil
		},
	}

	if y.IsComplete() {
		t.Error("expected Yield.IsComplete() to return false")
	}

	if !y.IsYielded() {
		t.Error("expected Yield.IsYielded() to return true")
	}
}

func TestResponse_Metadata(t *testing.T) {
	r := &chit.Response{
		Content:  "test",
		Metadata: map[string]any{"key": "value"},
	}

	if r.Metadata["key"] != "value" {
		t.Error("expected Response to store metadata")
	}
}

func TestYield_Metadata(t *testing.T) {
	y := &chit.Yield{
		Prompt:   "test",
		Metadata: map[string]any{"key": "value"},
		Continuation: func(_ context.Context, _ string, _ []chit.Message) (chit.Result, error) {
			return &chit.Response{Content: "continued"}, nil
		},
	}

	if y.Metadata["key"] != "value" {
		t.Error("expected Yield to store metadata")
	}
}

func TestYield_Continuation(t *testing.T) {
	continuationCalled := false

	y := &chit.Yield{
		Prompt: "test",
		Continuation: func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
			continuationCalled = true
			return &chit.Response{Content: "got: " + input}, nil
		},
	}

	result, err := y.Continuation(context.Background(), "hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !continuationCalled {
		t.Error("expected continuation to be called")
	}

	resp, ok := result.(*chit.Response)
	if !ok {
		t.Fatal("expected continuation to return Response")
	}

	if resp.Content != "got: hello" {
		t.Errorf("expected 'got: hello', got %q", resp.Content)
	}
}
