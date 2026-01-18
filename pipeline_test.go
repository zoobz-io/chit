package chit_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zoobzio/chit"
	helpers "github.com/zoobzio/chit/testing"
	"github.com/zoobzio/pipz"
)

func TestWithRetry_RetriesOnFailure(t *testing.T) {
	attempts := atomic.Int32{}
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		if attempts.Add(1) < 3 {
			return nil, errors.New("transient error")
		}
		return &chit.Response{Content: "success"}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter, chit.WithRetry(3))

	err := chat.Handle(context.Background(), "test")

	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}

	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}

	if emitter.Content() != "success" {
		t.Errorf("expected 'success', got %q", emitter.Content())
	}
}

func TestWithRetry_FailsAfterMaxAttempts(t *testing.T) {
	attempts := atomic.Int32{}
	expectedErr := errors.New("persistent error")
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		attempts.Add(1)
		return nil, expectedErr
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter, chit.WithRetry(3))

	err := chat.Handle(context.Background(), "test")

	if err == nil {
		t.Fatal("expected error after max retries")
	}

	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestWithTimeout_CancelsSlowProcessing(t *testing.T) {
	processor := chit.ProcessorFunc(func(ctx context.Context, input string, _ []chit.Message) (chit.Result, error) {
		select {
		case <-time.After(1 * time.Second):
			return &chit.Response{Content: "slow"}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter, chit.WithTimeout(50*time.Millisecond))

	err := chat.Handle(context.Background(), "test")

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWithMiddleware_RunsBeforeProcessor(t *testing.T) {
	order := make([]string, 0)

	middleware1 := pipz.Effect(
		pipz.NewIdentity("test:mw1", "First middleware"),
		func(_ context.Context, req *chit.ChatRequest) error {
			order = append(order, "middleware1")
			return nil
		},
	)

	middleware2 := pipz.Effect(
		pipz.NewIdentity("test:mw2", "Second middleware"),
		func(_ context.Context, req *chit.ChatRequest) error {
			order = append(order, "middleware2")
			return nil
		},
	)

	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		order = append(order, "processor")
		return &chit.Response{Content: "done"}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter,
		chit.WithMiddleware(middleware1, middleware2),
	)

	_ = chat.Handle(context.Background(), "test")

	if len(order) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(order))
	}

	if order[0] != "middleware1" || order[1] != "middleware2" || order[2] != "processor" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestWithMiddleware_CanModifyRequest(t *testing.T) {
	var modifiedInput string

	modifyMiddleware := pipz.Apply(
		pipz.NewIdentity("test:modify", "Modifies input"),
		func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			req.Input = "modified: " + req.Input
			modifiedInput = req.Input
			return req, nil
		},
	)

	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter,
		chit.WithMiddleware(modifyMiddleware),
	)

	_ = chat.Handle(context.Background(), "original")

	if modifiedInput != "modified: original" {
		t.Errorf("expected middleware to modify input, got %q", modifiedInput)
	}
}

func TestWithMiddleware_CanRejectRequest(t *testing.T) {
	rejectErr := errors.New("rejected")
	rejectMiddleware := pipz.Apply(
		pipz.NewIdentity("test:reject", "Rejects request"),
		func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			return nil, rejectErr
		},
	)

	processorCalled := false
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		processorCalled = true
		return &chit.Response{Content: "done"}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter,
		chit.WithMiddleware(rejectMiddleware),
	)

	err := chat.Handle(context.Background(), "test")

	if err == nil {
		t.Fatal("expected error from middleware")
	}

	if processorCalled {
		t.Error("processor should not be called when middleware rejects")
	}
}

func TestNewMiddleware_RejectsLongInput(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter,
		chit.WithMiddleware(chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			if len(req.Input) > 10 {
				return nil, errors.New("input too long")
			}
			return req, nil
		})),
	)

	err := chat.Handle(context.Background(), "this is too long")

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewMiddleware_RejectsEmptyInput(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter,
		chit.WithMiddleware(chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			if req.Input == "" {
				return nil, errors.New("input required")
			}
			return req, nil
		})),
	)

	err := chat.Handle(context.Background(), "")

	if err == nil {
		t.Fatal("expected validation error for empty input")
	}
}

func TestNewMiddleware_AcceptsValidInput(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter,
		chit.WithMiddleware(chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			return req, nil
		})),
	)

	err := chat.Handle(context.Background(), "hello")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetPipeline_ReturnsPipeline(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)

	pipeline := chat.GetPipeline()

	if pipeline == nil {
		t.Fatal("expected pipeline to be non-nil")
	}
}

func TestChatProvider_Interface(t *testing.T) {
	processor := helpers.EchoProcessor()
	emitter := &helpers.CollectingEmitter{}

	chat := chit.New(processor, emitter)

	// Verify Chat implements ChatProvider
	var _ chit.ChatProvider = chat
}

func TestMultiplePipelineOptions(t *testing.T) {
	attempts := atomic.Int32{}
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		if attempts.Add(1) < 2 {
			return nil, errors.New("transient")
		}
		return &chit.Response{Content: "success"}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter,
		chit.WithMiddleware(chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			if len(req.Input) > 100 {
				return nil, errors.New("input too long")
			}
			return req, nil
		})),
		chit.WithRetry(3),
		chit.WithTimeout(5*time.Second),
	)

	err := chat.Handle(context.Background(), "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if emitter.Content() != "success" {
		t.Errorf("expected 'success', got %q", emitter.Content())
	}
}

func TestContinuation_HasRetry(t *testing.T) {
	attempts := atomic.Int32{}

	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "What's your name?",
			Continuation: func(_ context.Context, name string, _ []chit.Message) (chit.Result, error) {
				if attempts.Add(1) < 3 {
					return nil, errors.New("transient continuation error")
				}
				return &chit.Response{Content: "Hello, " + name}, nil
			},
		}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter, chit.WithRetry(3))

	// First call yields
	err := chat.Handle(context.Background(), "start")
	if err != nil {
		t.Fatalf("unexpected error on yield: %v", err)
	}

	// Second call resumes - continuation fails twice, then succeeds
	err = chat.Handle(context.Background(), "Alice")
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}

	if attempts.Load() != 3 {
		t.Errorf("expected 3 continuation attempts, got %d", attempts.Load())
	}
}

func TestContinuation_HasTimeout(t *testing.T) {
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "What's your name?",
			Continuation: func(ctx context.Context, name string, _ []chit.Message) (chit.Result, error) {
				select {
				case <-time.After(1 * time.Second):
					return &chit.Response{Content: "slow"}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter, chit.WithTimeout(50*time.Millisecond))

	// First call yields
	_ = chat.Handle(context.Background(), "start")

	// Second call resumes - should timeout
	err := chat.Handle(context.Background(), "Alice")
	if err == nil {
		t.Fatal("expected timeout error on continuation")
	}
}

func TestContinuation_HasMiddleware(t *testing.T) {
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "Enter a short name:",
			Continuation: func(_ context.Context, name string, _ []chit.Message) (chit.Result, error) {
				return &chit.Response{Content: "Hello, " + name}, nil
			},
		}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter,
		chit.WithMiddleware(chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			if len(req.Input) > 5 {
				return nil, errors.New("input too long")
			}
			return req, nil
		})),
	)

	// First call yields
	_ = chat.Handle(context.Background(), "start")

	// Second call with too-long input should fail validation
	err := chat.Handle(context.Background(), "Alexander")
	if err == nil {
		t.Fatal("expected validation error on continuation input")
	}
}

func TestContinuation_ValidInputSucceeds(t *testing.T) {
	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "Enter a short name:",
			Continuation: func(_ context.Context, name string, _ []chit.Message) (chit.Result, error) {
				return &chit.Response{Content: "Hello, " + name}, nil
			},
		}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter,
		chit.WithMiddleware(chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
			if len(req.Input) > 10 {
				return nil, errors.New("input too long")
			}
			return req, nil
		})),
	)

	// First call yields
	_ = chat.Handle(context.Background(), "start")

	// Second call with valid input
	err := chat.Handle(context.Background(), "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check the response includes the continuation result
	content := emitter.Content()
	if content != "Enter a short name:Hello, Alice" {
		t.Errorf("expected 'Enter a short name:Hello, Alice', got %q", content)
	}
}

func TestContinuation_ReceivesUpdatedHistory(t *testing.T) {
	var continuationHistory []chit.Message

	processor := chit.ProcessorFunc(func(_ context.Context, input string, _ []chit.Message) (chit.Result, error) {
		return &chit.Yield{
			Prompt: "Continue?",
			Continuation: func(_ context.Context, input string, history []chit.Message) (chit.Result, error) {
				continuationHistory = history
				return &chit.Response{Content: "done"}, nil
			},
		}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	// First call - yields
	_ = chat.Handle(context.Background(), "first")

	// Second call - resumes
	_ = chat.Handle(context.Background(), "second")

	// Continuation should see: user "first", assistant "Continue?", user "second"
	if len(continuationHistory) != 3 {
		t.Fatalf("expected 3 messages in continuation history, got %d", len(continuationHistory))
	}

	if continuationHistory[0].Content != "first" {
		t.Errorf("expected first message 'first', got %q", continuationHistory[0].Content)
	}

	if continuationHistory[1].Content != "Continue?" {
		t.Errorf("expected second message 'Continue?', got %q", continuationHistory[1].Content)
	}

	if continuationHistory[2].Content != "second" {
		t.Errorf("expected third message 'second', got %q", continuationHistory[2].Content)
	}
}
