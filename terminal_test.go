package chit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/chit"
)

func TestNewTerminal_CallsProcessor(t *testing.T) {
	processorCalled := false
	var receivedInput string
	var receivedHistory []chit.Message

	processor := chit.ProcessorFunc(func(_ context.Context, input string, history []chit.Message) (chit.Result, error) {
		processorCalled = true
		receivedInput = input
		receivedHistory = history
		return &chit.Response{Content: "processed"}, nil
	})

	terminal := chit.NewTerminal(processor)
	history := []chit.Message{{Role: "user", Content: "previous"}}

	req := &chit.ChatRequest{
		Input:   "hello",
		History: history,
	}

	result, err := terminal.Process(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !processorCalled {
		t.Error("expected processor to be called")
	}

	if receivedInput != "hello" {
		t.Errorf("expected input 'hello', got %q", receivedInput)
	}

	if len(receivedHistory) != 1 || receivedHistory[0].Content != "previous" {
		t.Error("expected history to be passed through")
	}

	if result.Result == nil {
		t.Fatal("expected Result to be set")
	}

	resp, ok := result.Result.(*chit.Response)
	if !ok {
		t.Fatal("expected Result to be *Response")
	}

	if resp.Content != "processed" {
		t.Errorf("expected content 'processed', got %q", resp.Content)
	}
}

func TestNewTerminal_PropagatesError(t *testing.T) {
	expectedErr := errors.New("processor error")
	processor := chit.ProcessorFunc(func(_ context.Context, input string, history []chit.Message) (chit.Result, error) {
		return nil, expectedErr
	})

	terminal := chit.NewTerminal(processor)

	req := &chit.ChatRequest{
		Input:   "test",
		History: nil,
	}

	_, err := terminal.Process(context.Background(), req)

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewTerminal_HandlesYield(t *testing.T) {
	processor := chit.ProcessorFunc(func(_ context.Context, input string, history []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "need more info",
			Continuation: func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
				return &chit.Response{Content: "continued"}, nil
			},
		}, nil
	})

	terminal := chit.NewTerminal(processor)

	req := &chit.ChatRequest{
		Input:   "start",
		History: nil,
	}

	result, err := terminal.Process(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	yield, ok := result.Result.(*chit.Yield)
	if !ok {
		t.Fatal("expected Result to be *Yield")
	}

	if yield.Prompt != "need more info" {
		t.Errorf("expected prompt 'need more info', got %q", yield.Prompt)
	}

	if yield.Continuation == nil {
		t.Error("expected continuation to be set")
	}
}

func TestNewTerminal_HasIdentity(t *testing.T) {
	processor := chit.ProcessorFunc(func(_ context.Context, input string, history []chit.Message) (chit.Result, error) {
		return &chit.Response{Content: "test"}, nil
	})

	terminal := chit.NewTerminal(processor)

	identity := terminal.Identity()
	if identity.Name() != "chit:terminal" {
		t.Errorf("expected identity name 'chit:terminal', got %q", identity.Name())
	}
}

func TestNewTerminal_RespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	processor := chit.ProcessorFunc(func(ctx context.Context, input string, history []chit.Message) (chit.Result, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return &chit.Response{Content: "should not reach"}, nil
		}
	})

	terminal := chit.NewTerminal(processor)

	req := &chit.ChatRequest{
		Input:   "test",
		History: nil,
	}

	_, err := terminal.Process(ctx, req)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
