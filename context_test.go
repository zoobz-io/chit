package chit_test

import (
	"context"
	"testing"

	"github.com/zoobzio/chit"
	helpers "github.com/zoobzio/chit/testing"
)

func TestWithEmitter(t *testing.T) {
	emitter := &helpers.CollectingEmitter{}
	ctx := context.Background()

	ctx = chit.WithEmitter(ctx, emitter)

	retrieved := chit.EmitterFromContext(ctx)
	if retrieved != emitter {
		t.Error("expected to retrieve the same emitter from context")
	}
}

func TestEmitterFromContext_NotSet(t *testing.T) {
	ctx := context.Background()

	retrieved := chit.EmitterFromContext(ctx)
	if retrieved != nil {
		t.Error("expected nil when emitter not set in context")
	}
}

func TestHandle_InjectsEmitterIntoContext(t *testing.T) {
	var receivedEmitter chit.Emitter

	processor := chit.ProcessorFunc(func(ctx context.Context, _ string, _ []chit.Message) (chit.Result, error) {
		receivedEmitter = chit.EmitterFromContext(ctx)
		return &chit.Response{Content: "ok"}, nil
	})

	emitter := &helpers.CollectingEmitter{}
	chat := chit.New(processor, emitter)

	err := chat.Handle(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedEmitter != emitter {
		t.Error("expected processor to receive emitter via context")
	}
}
