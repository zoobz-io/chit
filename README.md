# chit

[![CI Status](https://github.com/zoobz-io/chit/workflows/CI/badge.svg)](https://github.com/zoobz-io/chit/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobz-io/chit/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobz-io/chit)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobz-io/chit)](https://goreportcard.com/report/github.com/zoobz-io/chit)
[![CodeQL](https://github.com/zoobz-io/chit/workflows/CodeQL/badge.svg)](https://github.com/zoobz-io/chit/security/code-scanning)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobz-io/chit.svg)](https://pkg.go.dev/github.com/zoobz-io/chit)
[![License](https://img.shields.io/github/license/zoobz-io/chit)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobz-io/chit)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobz-io/chit)](https://github.com/zoobz-io/chit/releases)

Conversation lifecycle controller for LLM-powered applications.

Manage user conversations, orchestrate pluggable processors, and stream responses — while keeping internal reasoning separate from what users see.

## The Conversation Belongs to You

Your processor receives input and read-only history. What it does with an LLM is its own business.

```go
// Processor handles reasoning — how it talks to LLMs is an implementation detail
processor := chit.ProcessorFunc(func(ctx context.Context, input string, history []chit.Message) (chit.Result, error) {
    // history is the user-facing conversation (read-only)
    // Internal LLM calls, chain-of-thought, tool use — none of it pollutes history
    response := callYourLLM(input, history)
    return &chit.Response{Content: response}, nil
})

// Emitter streams output to the user
emitter := &StreamingEmitter{writer: w}

// Chat manages the lifecycle
chat := chit.New(processor, emitter)

// Handle user input — chit manages history, you manage reasoning
chat.Handle(ctx, "What's the weather in Tokyo?")
```

The user sees a clean conversation. Your processor can have elaborate internal dialogues with the LLM — retries, tool calls, multi-step reasoning — and none of it leaks through.

## Install

```bash
go get github.com/zoobz-io/chit
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/zoobz-io/chit"
)

// SimpleEmitter collects messages for demonstration
type SimpleEmitter struct {
    Messages []chit.Message
}

func (e *SimpleEmitter) Emit(_ context.Context, msg chit.Message) error {
    e.Messages = append(e.Messages, msg)
    fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
    return nil
}
func (e *SimpleEmitter) Push(_ context.Context, _ chit.Resource) error { return nil }
func (e *SimpleEmitter) Close() error                                  { return nil }

func main() {
    // Processor returns responses or yields for multi-turn
    processor := chit.ProcessorFunc(func(_ context.Context, input string, history []chit.Message) (chit.Result, error) {
        // First message — ask for clarification
        if len(history) == 1 {
            return &chit.Yield{
                Prompt: "Which city would you like weather for?",
                Continuation: func(_ context.Context, city string, _ []chit.Message) (chit.Result, error) {
                    return &chit.Response{Content: fmt.Sprintf("Weather in %s: Sunny, 22°C", city)}, nil
                },
            }, nil
        }
        return &chit.Response{Content: "Hello! How can I help?"}, nil
    })

    emitter := &SimpleEmitter{}
    chat := chit.New(processor, emitter)

    // First call yields, asking for more info
    chat.Handle(context.Background(), "What's the weather?")
    // [assistant]: Which city would you like weather for?

    // Second call resumes with the answer
    chat.Handle(context.Background(), "Tokyo")
    // [assistant]: Weather in Tokyo: Sunny, 22°C

    // History tracked automatically
    fmt.Printf("Conversation has %d messages\n", len(chat.History()))
}
```

## Capabilities

| Feature | Description | Docs |
|---------|-------------|------|
| Processor Interface | Pluggable reasoning with read-only history | [Concepts](docs/1.learn/3.concepts.md) |
| Yield & Continue | Multi-turn conversations via continuations | [Concepts](docs/1.learn/3.concepts.md) |
| Pipeline Resilience | Retry, timeout, circuit breaker via [pipz](https://github.com/zoobz-io/pipz) | [Reliability](docs/2.guides/3.reliability.md) |
| Emitter Abstraction | Stream responses, push resources | [Architecture](docs/1.learn/4.architecture.md) |
| Signal Observability | Lifecycle events via [capitan](https://github.com/zoobz-io/capitan) | [Architecture](docs/1.learn/4.architecture.md) |
| Testing Utilities | Mock processors and emitters | [Testing](docs/2.guides/1.testing.md) |

## Why chit?

- **Clean separation** — User conversation stays clean; internal LLM reasoning is processor's business
- **Turn-taking built in** — Yield/Continue pattern for multi-turn without manual state management
- **Pipeline-native** — Wrap with [pipz](https://github.com/zoobz-io/pipz) for retry, timeout, rate limiting, circuit breakers
- **Observable** — Lifecycle signals via [capitan](https://github.com/zoobz-io/capitan) without instrumentation code
- **Bring your own LLM** — Processor interface works with any LLM client or framework

## Bring Your Own Reasoning

Chit manages the conversation lifecycle. How you reason is up to you.

```go
// Use zyn for typed LLM interactions
processor := chit.ProcessorFunc(func(ctx context.Context, input string, history []chit.Message) (chit.Result, error) {
    // Create internal session — separate from user history
    session := zyn.NewSession()
    session.Append(zyn.RoleSystem, "You are a helpful assistant.")

    // Add user context
    for _, msg := range history {
        session.Append(zyn.Role(msg.Role), msg.Content)
    }

    // Call LLM — internal retries, tool use, etc. stay internal
    response, _ := synapse.Process(ctx, session)
    return &chit.Response{Content: response}, nil
})

// Add resilience via pipz options
chat := chit.New(processor, emitter,
    chit.WithRetry(3),
    chit.WithTimeout(30*time.Second),
    chit.WithCircuitBreaker(5, time.Minute),
)
```

Your processor implementation can use [zyn](https://github.com/zoobz-io/zyn) for typed synapses, raw API calls, or any other approach. Chit doesn't care — it just manages what the user sees.

## Documentation

- **Learn**
  - [Overview](docs/1.learn/1.overview.md) — Purpose and design philosophy
  - [Quickstart](docs/1.learn/2.quickstart.md) — Get started in minutes
  - [Concepts](docs/1.learn/3.concepts.md) — Processors, results, emitters, history
  - [Architecture](docs/1.learn/4.architecture.md) — Internal design and pipeline integration
- **Guides**
  - [Testing](docs/2.guides/1.testing.md) — Mock processors and emitters
  - [Troubleshooting](docs/2.guides/2.troubleshooting.md) — Common issues and solutions
  - [Reliability](docs/2.guides/3.reliability.md) — Retry, timeout, circuit breakers
- **Reference**
  - [API](docs/4.reference/1.api.md) — Complete function documentation
  - [Types](docs/4.reference/2.types.md) — Message, Result, Response, Yield

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE) for details.
