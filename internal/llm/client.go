package llm

import "context"

// Role represents who sent a message
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a conversation message
type Message struct {
	Role       Role
	Content    string
	ToolCallID string     // For tool results
	ToolCalls  []ToolCall // For assistant messages with tool calls
}

// Tool describes a tool the LLM can call
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// ToolCall represents an LLM's request to call a tool
type ToolCall struct {
	ID         string
	Name       string
	Parameters map[string]interface{}
}

// Usage tracks token consumption
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Response from the LLM
type Response struct {
	Content   string     // Text response (if no tool call)
	ToolCalls []ToolCall // Tools the LLM wants to call
	Usage     Usage      // Token usage for this request
}

// Client is the interface all LLM providers implement
type Client interface {
	Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
}
