package chit

import (
	"time"

	"github.com/zoobz-io/pipz"
)

// Option is a functional option for configuring a Chat.
type Option func(*Chat)

// PipelineOption wraps a pipeline with additional behavior.
type PipelineOption func(pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest]

// Identities for pipeline options.
var (
	retryID          = pipz.NewIdentity("chit:retry", "Retries failed processing")
	backoffID        = pipz.NewIdentity("chit:backoff", "Retries with exponential backoff")
	timeoutID        = pipz.NewIdentity("chit:timeout", "Enforces processing timeout")
	circuitBreakerID = pipz.NewIdentity("chit:circuit-breaker", "Circuit breaker protection")
	rateLimitID      = pipz.NewIdentity("chit:rate-limit", "Rate limiting")
	fallbackID       = pipz.NewIdentity("chit:fallback", "Fallback processor")
	errorHandlerID   = pipz.NewIdentity("chit:error-handler", "Error handling")
	middlewareID     = pipz.NewIdentity("chit:middleware", "Middleware sequence")
)

// --- Configuration Options ---

// WithConfig sets the configuration for the Chat.
func WithConfig(config *Config) Option {
	return func(c *Chat) {
		c.config = config
	}
}

// WithSystemPrompt sets the system prompt for the Chat.
// This is a convenience option that creates a Config with the system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(c *Chat) {
		if c.config == nil {
			c.config = &Config{}
		}
		c.config.SystemPrompt = prompt
	}
}

// WithHistory sets the initial conversation history for the Chat.
func WithHistory(history []Message) Option {
	return func(c *Chat) {
		c.history = make([]Message, len(history))
		copy(c.history, history)
	}
}

// WithMetadata sets metadata on the Chat's configuration.
func WithMetadata(metadata map[string]any) Option {
	return func(c *Chat) {
		if c.config == nil {
			c.config = &Config{}
		}
		c.config.Metadata = metadata
	}
}

// --- Pipeline Options ---

// WithRetry adds retry logic to the pipeline.
// Failed requests are retried up to maxAttempts times.
func WithRetry(maxAttempts int) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewRetry(retryID, p, maxAttempts)
		})
	}
}

// WithBackoff adds retry logic with exponential backoff to the pipeline.
// Failed requests are retried with increasing delays between attempts.
func WithBackoff(maxAttempts int, baseDelay time.Duration) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewBackoff(backoffID, p, maxAttempts, baseDelay)
		})
	}
}

// WithTimeout adds timeout protection to the pipeline.
// Operations exceeding this duration will be canceled.
func WithTimeout(duration time.Duration) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewTimeout(timeoutID, p, duration)
		})
	}
}

// WithCircuitBreaker adds circuit breaker protection to the pipeline.
// After 'failures' consecutive failures, the circuit opens for 'recovery' duration.
func WithCircuitBreaker(failures int, recovery time.Duration) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewCircuitBreaker(circuitBreakerID, p, failures, recovery)
		})
	}
}

// WithRateLimit adds rate limiting to the pipeline.
// rps = requests per second, burst = burst capacity.
func WithRateLimit(rps float64, burst int) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewRateLimiter(rateLimitID, rps, burst, p)
		})
	}
}

// WithErrorHandler adds error handling to the pipeline.
// The error handler receives error context and can process/log/alert as needed.
func WithErrorHandler(handler pipz.Chainable[*pipz.Error[*ChatRequest]]) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewHandle(errorHandlerID, p, handler)
		})
	}
}

// ChatProvider is implemented by types that can provide a pipeline for composition.
type ChatProvider interface {
	GetPipeline() pipz.Chainable[*ChatRequest]
}

// WithFallback adds a fallback chat for resilience.
// If the primary processor fails, the fallback will be tried.
func WithFallback(fallback ChatProvider) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			return pipz.NewFallback(fallbackID, p, fallback.GetPipeline())
		})
	}
}

// WithMiddleware adds pre-processing steps before the terminal.
// Processors run in order, then the terminal (processor) runs.
func WithMiddleware(processors ...pipz.Chainable[*ChatRequest]) Option {
	return func(c *Chat) {
		c.pipelineOpts = append(c.pipelineOpts, func(p pipz.Chainable[*ChatRequest]) pipz.Chainable[*ChatRequest] {
			steps := make([]pipz.Chainable[*ChatRequest], 0, len(processors)+1)
			steps = append(steps, processors...)
			steps = append(steps, p)
			return pipz.NewSequence(middlewareID, steps...)
		})
	}
}
