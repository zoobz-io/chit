package chit

// ChatRequest flows through the pipz pipeline.
// It contains input, conversation history (read-only), metadata, and output fields.
type ChatRequest struct {
	// Input fields
	Input   string    // User input to process
	History []Message // User conversation history (read-only snapshot)

	// Metadata fields
	ChatID    string // ID of the chat instance
	RequestID string // Unique identifier for this request

	// Output fields (populated by terminal)
	Result Result // Processing result (Response or Yield)
}

// Clone implements pipz.Cloner for concurrent middleware support.
func (r *ChatRequest) Clone() *ChatRequest {
	history := make([]Message, len(r.History))
	copy(history, r.History)

	return &ChatRequest{
		Input:     r.Input,
		History:   history,
		ChatID:    r.ChatID,
		RequestID: r.RequestID,
		Result:    r.Result,
	}
}
