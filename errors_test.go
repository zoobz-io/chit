package chit_test

import (
	"errors"
	"testing"

	"github.com/zoobz-io/chit"
)

func TestErrUnknownResultType(t *testing.T) {
	err := chit.ErrUnknownResultType

	if err == nil {
		t.Fatal("expected ErrUnknownResultType to be defined")
	}

	if err.Error() != "chit: unknown result type" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestErrNilProcessor(t *testing.T) {
	err := chit.ErrNilProcessor

	if err == nil {
		t.Fatal("expected ErrNilProcessor to be defined")
	}

	if err.Error() != "chit: processor is required" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestErrNilEmitter(t *testing.T) {
	err := chit.ErrNilEmitter

	if err == nil {
		t.Fatal("expected ErrNilEmitter to be defined")
	}

	if err.Error() != "chit: emitter is required" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestErrEmitterClosed(t *testing.T) {
	err := chit.ErrEmitterClosed

	if err == nil {
		t.Fatal("expected ErrEmitterClosed to be defined")
	}

	if err.Error() != "chit: emitter is closed" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestErrors_AreDistinct(t *testing.T) {
	errs := []error{
		chit.ErrUnknownResultType,
		chit.ErrNilProcessor,
		chit.ErrNilEmitter,
		chit.ErrEmitterClosed,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("errors should be distinct: %v and %v", err1, err2)
			}
		}
	}
}

func TestErrors_CanBeWrapped(t *testing.T) {
	wrapped := errors.Join(errors.New("context"), chit.ErrUnknownResultType)

	if !errors.Is(wrapped, chit.ErrUnknownResultType) {
		t.Error("expected wrapped error to match ErrUnknownResultType")
	}
}
