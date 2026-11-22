package llm

import (
	"context"
	"strings"

	"encoding/json/jsontext"
)

// Role says who wrote a message.
type Role string

// Roles in a conversation. Tool results are sent back with RoleUser:
// from the model's point of view, "the user" is whoever runs its tools.
const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// BlockType says what kind of content a ContentBlock carries.
type BlockType string

// Content block types of the neutral format.
const (
	BlockText       BlockType = "text"        // the model (or user) is talking
	BlockToolUse    BlockType = "tool_use"    // the model asks us to run a tool
	BlockToolResult BlockType = "tool_result" // we send back what the tool returned
)

// StopReason says why the model stopped generating.
type StopReason string

// Stop reasons the agent loop must handle.
const (
	StopEndTurn   StopReason = "end_turn"   // finished: the answer is ready
	StopToolUse   StopReason = "tool_use"   // waiting for tool results
	StopMaxTokens StopReason = "max_tokens" // cut off by MaxTokens: the answer is incomplete
)

// ContentBlock is one piece of a message. Which fields are set depends on Type:
//
//	text:        Text
//	tool_use:    ID, Name, Input
//	tool_result: ToolUseID, Content, IsError
//
// Fields that don't belong to the block's Type must be left out of the JSON
// entirely, not sent as "" or false.
type ContentBlock struct {
	// TestRequestMarshalGolden and TestResponseUnmarshal pass.
	// Look at testdata/*.json for the exact field names.
	// Hint: `json:"name,omitzero"` leaves a field out when it holds its zero value.
	Type BlockType `json:"type"`
	Text string    `json:"text,omitzero"`

	// tool_use
	ID    string         `json:"id,omitzero"`
	Name  string         `json:"name,omitzero"`
	Input jsontext.Value `json:"input,omitzero"` // tool input: raw JSON, its shape depends on the tool

	// tool_result
	ToolUseID string `json:"tool_use_id,omitzero"`
	Content   string `json:"content,omitzero"`
	IsError   bool   `json:"is_error,omitzero"`
}

// Message is one turn of the conversation.
type Message struct {
	Role    Role           `json:"role"`
	Content []ContentBlock `json:"content"`
}

// Tool describes a tool to the model. The model reads only this, never our Go code.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema jsontext.Value `json:"input_schema"`
}

// Request is everything we send to the model in one call. The API is stateless,
// so Messages must contain the whole conversation every time.
type Request struct {
	Model     string    `json:"model,omitzero"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitzero"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitzero"`
}

// Usage counts tokens: the unit providers bill by and context limits are measured in.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Response is the model's reply to one Request.
type Response struct {
	Content    []ContentBlock `json:"content"`
	StopReason StopReason     `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
}

// Text returns all text blocks of the response joined together, skipping
// tool_use and any other non-text blocks. Never assume Content[0] is text:
// the first block may be a tool call.
func (r *Response) Text() string {

	var sb strings.Builder

	for _, content := range r.Content {
		if content.Type == BlockText {
			sb.WriteString(content.Text)
		}
	}

	return sb.String()
}

// Provider is the one thing the agent knows about language models. Every
// adapter (Ollama, Claude, OpenAI-compatible) implements it, and so do
// decorators such as Fallback and retry wrappers.
type Provider interface {
	Send(ctx context.Context, req Request) (*Response, error)
}
