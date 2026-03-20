package chit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/chit"
)

func TestNewMiddleware_TransformsRequest(t *testing.T) {
	middleware := chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
		req.Input = "transformed: " + req.Input
		return req, nil
	})

	req := &chit.ChatRequest{
		Input:   "original",
		History: nil,
	}

	result, err := middleware.Process(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input != "transformed: original" {
		t.Errorf("expected 'transformed: original', got %q", result.Input)
	}
}

func TestNewMiddleware_ReturnsError(t *testing.T) {
	middleware := chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
		if req.Input == "" {
			return nil, errors.New("input required")
		}
		return req, nil
	})

	req := &chit.ChatRequest{
		Input:   "",
		History: nil,
	}

	_, err := middleware.Process(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestNewMiddleware_PassThrough(t *testing.T) {
	middleware := chit.NewMiddleware(func(_ context.Context, req *chit.ChatRequest) (*chit.ChatRequest, error) {
		return req, nil
	})

	req := &chit.ChatRequest{
		Input:   "unchanged",
		History: nil,
	}

	result, err := middleware.Process(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input != "unchanged" {
		t.Errorf("expected 'unchanged', got %q", result.Input)
	}
}
