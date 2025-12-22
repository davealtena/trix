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

const anthropicAPI = "https://api.anthropic.com/v1/messages"

// AnthropicClient implements the client interface for Claude
type AnthropicClient struct {
	apikey string
	model  string
	client *http.Client
}

// NewAnthropicClient creates a new Claude client
func NewAnthropicClient(model string) (*AnthropicClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return &AnthropicClient{
		apikey: apiKey,
		model:  model,
		client: &http.Client{},
	}, nil
}

// Chat sends messages to Claude and returns the reponse
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {

	// Convert to Anthropic format
	reqBody := c.buildRequest(messages, tools)

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPI, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create requests: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apikey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return c.parseResponseBody(body)
}

func (c *AnthropicClient) buildRequest(messages []Message, tools []Tool) map[string]interface{} {
	// Separate system message from conversation
	var system string
	var convMessages []map[string]interface{}

	for _, msg := range messages {
		if msg.Role == RoleSystem {
			system = msg.Content
			continue
		}

		anthropicMsg := map[string]interface{}{
			"role": string(msg.Role),
		}

		if msg.Role == RoleTool {
			// Tool result format
			anthropicMsg["role"] = "user"
			anthropicMsg["content"] = []map[string]interface{}{
				{
					"type":        "tool_result",
					"tool_use_id": msg.ToolCallID,
					"content":     msg.Content,
				},
			}
		} else if msg.Role == RoleAssistant && len(msg.ToolCalls) > 0 {
			// Assistant message with tool calls
			var contentBlocks []map[string]interface{}
			if msg.Content != "" {
				contentBlocks = append(contentBlocks, map[string]interface{}{
					"type": "text",
					"text": msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				contentBlocks = append(contentBlocks, map[string]interface{}{
					"type":  "tool_use",
					"id":    tc.ID,
					"name":  tc.Name,
					"input": tc.Parameters,
				})
			}
			anthropicMsg["content"] = contentBlocks
		} else {
			anthropicMsg["content"] = msg.Content
		}

		convMessages = append(convMessages, anthropicMsg)
	}

	req := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"messages":   convMessages,
	}

	if system != "" {
		req["system"] = system
	}

	if len(tools) > 0 {
		req["tools"] = c.convertTools(tools)
	}

	return req
}

func (c *AnthropicClient) convertTools(tools []Tool) []map[string]interface{} {
	var result []map[string]interface{}
	for _, tool := range tools {
		result = append(result, map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.Parameters,
		})
	}
	return result
}

func (c *AnthropicClient) parseResponseBody(body []byte) (*Response, error) {
	var apiResp struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text,omitempty"`
			ID    string          `json:"id,omitempty"`
			Name  string          `json:"name,omitempty"`
			Input json.RawMessage `json:"input,omitempty"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := &Response{
		Usage: Usage{
			InputTokens:  apiResp.Usage.InputTokens,
			OutputTokens: apiResp.Usage.OutputTokens,
		},
	}

	for _, block := range apiResp.Content {
		switch block.Type {
		case "text":
			response.Content += block.Text
		case "tool_use":
			var params map[string]interface{}
			if err := json.Unmarshal(block.Input, &params); err != nil {
				params = make(map[string]interface{})
			}
			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID:         block.ID,
				Name:       block.Name,
				Parameters: params,
			})
		}
	}
	return response, nil
}
