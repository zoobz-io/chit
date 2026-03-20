package chit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/chit"
)

func TestProcessorFunc(t *testing.T) {
	called := false
	var receivedInput string
	var receivedHistory []chit.Message

	fn := chit.ProcessorFunc(func(ctx context.Context, input string, history []chit.Message) (chit.Result, error) {
		called = true
		receivedInput = input
		receivedHistory = history
		return &chit.Response{Content: "processed: " + input}, nil
	})

	history := []chit.Message{{Role: "user", Content: "previous"}}
	result, err := fn.Process(context.Background(), "hello", history)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("expected ProcessorFunc to be called")
	}

	if receivedInput != "hello" {
		t.Errorf("expected input 'hello', got %q", receivedInput)
	}

	if len(receivedHistory) != 1 || receivedHistory[0].Content != "previous" {
		t.Error("expected history to be passed through")
	}

	resp, ok := result.(*chit.Response)
	if !ok {
		t.Fatal("expected Response result")
	}

	if resp.Content != "processed: hello" {
		t.Errorf("expected 'processed: hello', got %q", resp.Content)
	}
}

func TestProcessorFunc_ReturnsError(t *testing.T) {
	expectedErr := errors.New("processor error")

	fn := chit.ProcessorFunc(func(_ context.Context, _ string, _ []chit.Message) (chit.Result, error) {
		return nil, expectedErr
	})

	result, err := fn.Process(context.Background(), "test", nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestProcessorFunc_ReturnsYield(t *testing.T) {
	fn := chit.ProcessorFunc(func(_ context.Context, _ string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "need more info",
			Continuation: func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
				return &chit.Response{Content: "got: " + input}, nil
			},
		}, nil
	})

	result, err := fn.Process(context.Background(), "start", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	yield, ok := result.(*chit.Yield)
	if !ok {
		t.Fatal("expected Yield result")
	}

	if yield.Prompt != "need more info" {
		t.Errorf("expected prompt 'need more info', got %q", yield.Prompt)
	}

	if yield.Continuation == nil {
		t.Error("expected continuation to be set")
	}
}
