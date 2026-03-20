package testing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/chit"
	helpers "github.com/zoobz-io/chit/testing"
)

func TestMockProcessor_DefaultBehavior(t *testing.T) {
	mock := &helpers.MockProcessor{}

	result, err := mock.Process(context.Background(), "test", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, ok := result.(*chit.Response)
	if !ok {
		t.Fatal("expected Response")
	}

	if resp.Content != "" {
		t.Errorf("expected empty content, got %q", resp.Content)
	}
}

func TestMockProcessor_CustomFunc(t *testing.T) {
	mock := &helpers.MockProcessor{
		ProcessFunc: func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
			return &chit.Response{Content: "custom: " + input}, nil
		},
	}

	result, err := mock.Process(context.Background(), "hello", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(*chit.Response)
	if resp.Content != "custom: hello" {
		t.Errorf("expected 'custom: hello', got %q", resp.Content)
	}
}

func TestMockProcessor_RecordsCalls(t *testing.T) {
	mock := &helpers.MockProcessor{}
	history := []chit.Message{{Role: "user", Content: "previous"}}

	_, _ = mock.Process(context.Background(), "first", history)
	_, _ = mock.Process(context.Background(), "second", history)

	if len(mock.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.Calls))
	}

	if mock.Calls[0].Input != "first" {
		t.Errorf("expected first call input 'first', got %q", mock.Calls[0].Input)
	}

	if mock.Calls[1].Input != "second" {
		t.Errorf("expected second call input 'second', got %q", mock.Calls[1].Input)
	}

	if len(mock.Calls[0].History) != 1 {
		t.Error("expected history to be recorded")
	}
}

func TestMockProcessor_ReturnsError(t *testing.T) {
	expectedErr := errors.New("mock error")
	mock := &helpers.MockProcessor{
		ProcessFunc: func(_ context.Context, _ string, _ []chit.Message) (chit.Result, error) {
			return nil, expectedErr
		},
	}

	result, err := mock.Process(context.Background(), "test", nil)

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestMockEmitter_DefaultBehavior(t *testing.T) {
	mock := &helpers.MockEmitter{}

	err := mock.Emit(context.Background(), chit.Message{Content: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = mock.Push(context.Background(), chit.Resource{Type: "data"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(mock.Messages))
	}

	if len(mock.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(mock.Resources))
	}

	if !mock.Closed {
		t.Error("expected Closed to be true")
	}
}

func TestMockEmitter_CustomFuncs(t *testing.T) {
	emitCalled := false
	pushCalled := false
	closeCalled := false

	mock := &helpers.MockEmitter{
		EmitFunc: func(_ context.Context, _ chit.Message) error {
			emitCalled = true
			return nil
		},
		PushFunc: func(_ context.Context, _ chit.Resource) error {
			pushCalled = true
			return nil
		},
		CloseFunc: func() error {
			closeCalled = true
			return nil
		},
	}

	_ = mock.Emit(context.Background(), chit.Message{})
	_ = mock.Push(context.Background(), chit.Resource{})
	_ = mock.Close()

	if !emitCalled {
		t.Error("expected EmitFunc to be called")
	}
	if !pushCalled {
		t.Error("expected PushFunc to be called")
	}
	if !closeCalled {
		t.Error("expected CloseFunc to be called")
	}
}

func TestMockEmitter_ReturnsErrors(t *testing.T) {
	emitErr := errors.New("emit error")
	pushErr := errors.New("push error")
	closeErr := errors.New("close error")

	mock := &helpers.MockEmitter{
		EmitFunc:  func(_ context.Context, _ chit.Message) error { return emitErr },
		PushFunc:  func(_ context.Context, _ chit.Resource) error { return pushErr },
		CloseFunc: func() error { return closeErr },
	}

	if err := mock.Emit(context.Background(), chit.Message{}); !errors.Is(err, emitErr) {
		t.Errorf("expected emit error")
	}

	if err := mock.Push(context.Background(), chit.Resource{}); !errors.Is(err, pushErr) {
		t.Errorf("expected push error")
	}

	if err := mock.Close(); !errors.Is(err, closeErr) {
		t.Errorf("expected close error")
	}
}

func TestCollectingEmitter_CollectsMessages(t *testing.T) {
	emitter := &helpers.CollectingEmitter{}

	_ = emitter.Emit(context.Background(), chit.Message{Role: "assistant", Content: "Hello"})
	_ = emitter.Emit(context.Background(), chit.Message{Role: "assistant", Content: " World"})

	if len(emitter.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(emitter.Messages))
	}

	if emitter.Content() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", emitter.Content())
	}
}

func TestCollectingEmitter_CollectsResources(t *testing.T) {
	emitter := &helpers.CollectingEmitter{}

	_ = emitter.Push(context.Background(), chit.Resource{Type: "data", URI: "db://users/1"})
	_ = emitter.Push(context.Background(), chit.Resource{Type: "context", URI: "db://users/2"})

	if len(emitter.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(emitter.Resources))
	}

	if emitter.Resources[0].URI != "db://users/1" {
		t.Errorf("expected first URI 'db://users/1', got %q", emitter.Resources[0].URI)
	}
}

func TestCollectingEmitter_CloseIsNoop(t *testing.T) {
	emitter := &helpers.CollectingEmitter{}

	err := emitter.Close()
	if err != nil {
		t.Errorf("expected Close to return nil, got %v", err)
	}
}

func TestEchoProcessor(t *testing.T) {
	processor := helpers.EchoProcessor()

	result, err := processor.Process(context.Background(), "hello", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, ok := result.(*chit.Response)
	if !ok {
		t.Fatal("expected Response")
	}

	if resp.Content != "Echo: hello" {
		t.Errorf("expected 'Echo: hello', got %q", resp.Content)
	}
}

func TestYieldingProcessor_FirstCallYields(t *testing.T) {
	processor := helpers.YieldingProcessor("What's your name?", "Hello")

	result, err := processor.Process(context.Background(), "start", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	yield, ok := result.(*chit.Yield)
	if !ok {
		t.Fatal("expected Yield on first call")
	}

	if yield.Prompt != "What's your name?" {
		t.Errorf("expected prompt 'What's your name?', got %q", yield.Prompt)
	}

	if yield.Continuation == nil {
		t.Fatal("expected Continuation to be set")
	}
}

func TestYieldingProcessor_ContinuationResponds(t *testing.T) {
	processor := helpers.YieldingProcessor("What's your name?", "Hello")

	result, _ := processor.Process(context.Background(), "start", nil)
	yield := result.(*chit.Yield)

	// Call continuation
	resumed, err := yield.Continuation(context.Background(), "Alice", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, ok := resumed.(*chit.Response)
	if !ok {
		t.Fatal("expected Response from continuation")
	}

	if resp.Content != "Hello: Alice" {
		t.Errorf("expected 'Hello: Alice', got %q", resp.Content)
	}
}
