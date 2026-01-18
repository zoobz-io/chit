package chit

import "github.com/zoobzio/capitan"

// Signal definitions for chit conversation lifecycle events.
// Signals follow the pattern: chit.<entity>.<event>.
var (
	// Chat lifecycle signals.
	ChatCreated = capitan.NewSignal(
		"chit.chat.created",
		"New chat conversation initiated",
	)

	// Input signals.
	InputReceived = capitan.NewSignal(
		"chit.input.received",
		"User input received for processing",
	)

	// Processing signals.
	ProcessingStarted = capitan.NewSignal(
		"chit.processing.started",
		"Processor began handling input",
	)
	ProcessingCompleted = capitan.NewSignal(
		"chit.processing.completed",
		"Processor finished handling input",
	)
	ProcessingFailed = capitan.NewSignal(
		"chit.processing.failed",
		"Processor encountered an error",
	)

	// Response signals.
	ResponseEmitted = capitan.NewSignal(
		"chit.response.emitted",
		"Response message emitted to client",
	)

	// Resource signals.
	ResourcePushed = capitan.NewSignal(
		"chit.resource.pushed",
		"Structured resource pushed to client",
	)

	// Turn management signals.
	TurnYielded = capitan.NewSignal(
		"chit.turn.yielded",
		"Processing yielded awaiting input",
	)
	TurnResumed = capitan.NewSignal(
		"chit.turn.resumed",
		"Processing resumed from continuation",
	)
)

// Field keys for chit event data.
var (
	// Chat metadata.
	FieldChatID = capitan.NewStringKey("chat_id")

	// Input metadata.
	FieldInput     = capitan.NewStringKey("input")
	FieldInputSize = capitan.NewIntKey("input_size")

	// Processing metadata.
	FieldProcessingDuration = capitan.NewDurationKey("processing_duration")

	// Response metadata.
	FieldRole        = capitan.NewStringKey("role")
	FieldContentSize = capitan.NewIntKey("content_size")

	// Resource metadata.
	FieldResourceType = capitan.NewStringKey("resource_type")
	FieldResourceURI  = capitan.NewStringKey("resource_uri")

	// Turn metadata.
	FieldPrompt = capitan.NewStringKey("prompt")

	// Error information.
	FieldError = capitan.NewErrorKey("error")
)
