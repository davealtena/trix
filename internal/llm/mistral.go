package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const mistralAPIURL = "https://api.mistral.ai/v1/chat/completions"

// MistralClient implements the Client interface for Mistral AI.
type MistralClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewMistralClient creates a new Mistral client.
// Reads API key from MISTRAL_API_KEY environment variable.
func NewMistralClient(model string) (*MistralClient, error) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY environment variable not set")
	}

	if model == "" {
		model = "mistral-large-latest"
	}

	return &MistralClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

// Mistral API request/response types
type mistralRequest struct {
	Model       string           `json:"model"`
	Messages    []mistralMessage `json:"messages"`
	Tools       []mistralTool    `json:"tools,omitempty"`
	ToolChoice  string           `json:"tool_choice,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
}

type mistralMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	ToolCalls  []mistralToolCall `json:"tool_calls,omitempty"`
}

type mistralToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function mistralFunctionCall `json:"function"`
}

type mistralFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type mistralTool struct {
	Type     string          `json:"type"`
	Function mistralFunction `json:"function"`
}

type mistralFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type mistralResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string            `json:"role"`
			Content   string            `json:"content"`
			ToolCalls []mistralToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type mistralError struct {
	Object  string `json:"object"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Chat sends messages to Mistral and returns the response.
func (c *MistralClient) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	req := mistralRequest{
		Model:       c.model,
		Messages:    c.convertMessages(messages),
		Temperature: 0.7,
		TopP:        1.0,
		MaxTokens:   4096,
	}

	if len(tools) > 0 {
		req.Tools = c.convertTools(tools)
		req.ToolChoice = "auto"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", mistralAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr mistralError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("(HTTP Error %d) %s", resp.StatusCode, apiErr.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var mistralResp mistralResponse
	if err := json.Unmarshal(respBody, &mistralResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return c.parseResponse(&mistralResp), nil
}

// convertMessages converts generic Messages to Mistral's format.
func (c *MistralClient) convertMessages(messages []Message) []mistralMessage {
	var result []mistralMessage

	for _, msg := range messages {
		mistralMsg := mistralMessage{
			Content: msg.Content,
		}

		switch msg.Role {
		case RoleSystem:
			mistralMsg.Role = "system"
		case RoleUser:
			mistralMsg.Role = "user"
		case RoleAssistant:
			mistralMsg.Role = "assistant"
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					argsJSON, _ := json.Marshal(tc.Parameters)
					mistralMsg.ToolCalls = append(mistralMsg.ToolCalls, mistralToolCall{
						ID:   tc.ID,
						Type: "function",
						Function: mistralFunctionCall{
							Name:      tc.Name,
							Arguments: string(argsJSON),
						},
					})
				}
			}
		case RoleTool:
			mistralMsg.Role = "tool"
			mistralMsg.ToolCallID = msg.ToolCallID
		}

		result = append(result, mistralMsg)
	}

	return result
}

// convertTools converts generic Tools to Mistral's format.
func (c *MistralClient) convertTools(tools []Tool) []mistralTool {
	var result []mistralTool

	for _, tool := range tools {
		result = append(result, mistralTool{
			Type: "function",
			Function: mistralFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}

	return result
}

// parseResponse converts Mistral's response to the generic Response type.
func (c *MistralClient) parseResponse(resp *mistralResponse) *Response {
	if len(resp.Choices) == 0 {
		return &Response{}
	}

	choice := resp.Choices[0]
	response := &Response{
		Content: choice.Message.Content,
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	for _, tc := range choice.Message.ToolCalls {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
			params = make(map[string]interface{})
		}

		response.ToolCalls = append(response.ToolCalls, ToolCall{
			ID:         tc.ID,
			Name:       tc.Function.Name,
			Parameters: params,
		})
	}

	return response
}
