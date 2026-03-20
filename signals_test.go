package chit_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/chit"
	helpers "github.com/zoobz-io/chit/testing"
)

// waitFor waits for a condition to be true, with timeout.
func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if cond() || time.Now().After(deadline) {
			return cond()
		}
		time.Sleep(time.Millisecond)
	}
}

func TestChatCreatedSignal(t *testing.T) {
	var mu sync.Mutex
	var receivedChatID string

	listener := capitan.Hook(chit.ChatCreated, func(_ context.Context, e *capitan.Event) {
		chatID, _ := chit.FieldChatID.From(e)
		mu.Lock()
		receivedChatID = chatID
		mu.Unlock()
	})
	defer listener.Close()

	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedChatID != ""
	}) {
		t.Fatal("expected ChatCreated signal")
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedChatID != chat.ID() {
		t.Errorf("expected chat_id %q, got %q", chat.ID(), receivedChatID)
	}
}

func TestInputReceivedSignal(t *testing.T) {
	var mu sync.Mutex
	var receivedInput string
	var receivedSize int

	listener := capitan.Hook(chit.InputReceived, func(_ context.Context, e *capitan.Event) {
		input, _ := chit.FieldInput.From(e)
		size, _ := chit.FieldInputSize.From(e)
		mu.Lock()
		receivedInput = input
		receivedSize = size
		mu.Unlock()
	})
	defer listener.Close()

	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "Hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedInput != ""
	}) {
		t.Fatal("expected InputReceived signal")
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedInput != "Hello world" {
		t.Errorf("expected input 'Hello world', got %q", receivedInput)
	}
	if receivedSize != len("Hello world") {
		t.Errorf("expected input_size %d, got %d", len("Hello world"), receivedSize)
	}
}

func TestProcessingSignals(t *testing.T) {
	var mu sync.Mutex
	var started, completed bool
	var duration time.Duration

	startListener := capitan.Hook(chit.ProcessingStarted, func(_ context.Context, _ *capitan.Event) {
		mu.Lock()
		started = true
		mu.Unlock()
	})
	defer startListener.Close()

	completeListener := capitan.Hook(chit.ProcessingCompleted, func(_ context.Context, e *capitan.Event) {
		dur, _ := chit.FieldProcessingDuration.From(e)
		mu.Lock()
		completed = true
		duration = dur
		mu.Unlock()
	})
	defer completeListener.Close()

	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return started && completed
	}) {
		t.Fatal("expected ProcessingStarted and ProcessingCompleted signals")
	}

	mu.Lock()
	defer mu.Unlock()

	if duration <= 0 {
		t.Errorf("expected positive duration, got %v", duration)
	}
}

func TestProcessingFailedSignal(t *testing.T) {
	var mu sync.Mutex
	var receivedErr error

	listener := capitan.Hook(chit.ProcessingFailed, func(_ context.Context, e *capitan.Event) {
		err, _ := chit.FieldError.From(e)
		mu.Lock()
		receivedErr = err
		mu.Unlock()
	})
	defer listener.Close()

	expectedErr := errors.New("processor failed")
	processor := chit.ProcessorFunc(func(_ context.Context, _ string, _ []chit.Message) (chit.Result, error) {
		return nil, expectedErr
	})
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "test")
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedErr != nil
	}) {
		t.Fatal("expected ProcessingFailed signal")
	}

	mu.Lock()
	defer mu.Unlock()

	if !errors.Is(receivedErr, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, receivedErr)
	}
}

func TestResponseEmittedSignal(t *testing.T) {
	var mu sync.Mutex
	var receivedRole string
	var receivedSize int

	listener := capitan.Hook(chit.ResponseEmitted, func(_ context.Context, e *capitan.Event) {
		role, _ := chit.FieldRole.From(e)
		size, _ := chit.FieldContentSize.From(e)
		mu.Lock()
		receivedRole = role
		receivedSize = size
		mu.Unlock()
	})
	defer listener.Close()

	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedRole != ""
	}) {
		t.Fatal("expected ResponseEmitted signal")
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedRole != "assistant" {
		t.Errorf("expected role 'assistant', got %q", receivedRole)
	}
	// "Echo: Hello" = 11 chars
	if receivedSize != 11 {
		t.Errorf("expected content_size 11, got %d", receivedSize)
	}
}

func TestTurnYieldedSignal(t *testing.T) {
	var mu sync.Mutex
	var receivedPrompt string

	listener := capitan.Hook(chit.TurnYielded, func(_ context.Context, e *capitan.Event) {
		prompt, _ := chit.FieldPrompt.From(e)
		mu.Lock()
		receivedPrompt = prompt
		mu.Unlock()
	})
	defer listener.Close()

	processor := helpers.YieldingProcessor("What is your name?", "Hello")
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "Start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedPrompt != ""
	}) {
		t.Fatal("expected TurnYielded signal")
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedPrompt != "What is your name?" {
		t.Errorf("expected prompt 'What is your name?', got %q", receivedPrompt)
	}
}

func TestTurnResumedSignal(t *testing.T) {
	var mu sync.Mutex
	var resumed bool

	listener := capitan.Hook(chit.TurnResumed, func(_ context.Context, _ *capitan.Event) {
		mu.Lock()
		resumed = true
		mu.Unlock()
	})
	defer listener.Close()

	processor := helpers.YieldingProcessor("What is your name?", "Hello")
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	// First call yields
	err := chat.Handle(context.Background(), "Start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not have resumed yet
	mu.Lock()
	if resumed {
		mu.Unlock()
		t.Fatal("TurnResumed should not fire on first call")
	}
	mu.Unlock()

	// Second call resumes
	err = chat.Handle(context.Background(), "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return resumed
	}) {
		t.Fatal("expected TurnResumed signal")
	}
}

func TestSignalChatIDCorrelation(t *testing.T) {
	var mu sync.Mutex
	chatIDs := make(map[string]int)

	signals := []capitan.Signal{
		chit.ChatCreated,
		chit.InputReceived,
		chit.ProcessingStarted,
		chit.ProcessingCompleted,
		chit.ResponseEmitted,
	}

	listeners := make([]*capitan.Listener, 0, len(signals))
	for _, sig := range signals {
		listener := capitan.Hook(sig, func(_ context.Context, e *capitan.Event) {
			if chatID, ok := chit.FieldChatID.From(e); ok {
				mu.Lock()
				chatIDs[chatID]++
				mu.Unlock()
			}
		})
		listeners = append(listeners, listener)
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for all signals (5 expected)
	if !waitFor(time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		total := 0
		for _, count := range chatIDs {
			total += count
		}
		return total >= 5
	}) {
		t.Fatal("expected at least 5 signals")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(chatIDs) != 1 {
		t.Errorf("expected all signals to share one chat_id, got %d unique: %v", len(chatIDs), chatIDs)
	}

	for chatID := range chatIDs {
		if chatID != chat.ID() {
			t.Errorf("signal chat_id %q doesn't match chat.ID() %q", chatID, chat.ID())
		}
	}
}
