package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go" // imported as openai
)

type OpenAIClient struct {
	client openai.Client
}

func NewOpenAIClient() (*OpenAIClient, error) {
	return &OpenAIClient{
		client: openai.NewClient(),
	}, nil
}

// Chat sends messages to OpenAI and returns the response
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	// ===== STEP 1: Convert your Messages to OpenAI format =====
	var openaiMessages []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		switch msg.Role {
		case RoleSystem:
			// System message: instructions for the AI
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))

		case RoleUser:
			// User message: what the user is asking
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))

		case RoleAssistant:
			// Assistant message: previous AI response
			// If the assistant made tool calls, we need to include those
			if len(msg.ToolCalls) > 0 {
				// Build tool calls array for OpenAI
				var toolCalls []openai.ChatCompletionMessageToolCallParam
				for _, tc := range msg.ToolCalls {
					// Convert parameters to JSON string (OpenAI expects raw JSON)
					argsJSON, _ := json.Marshal(tc.Parameters)
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID:   tc.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: string(argsJSON),
						},
					})
				}
				// Build assistant message with tool calls directly
				openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: toolCalls,
					},
				})
			} else {
				openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
			}

		case RoleTool:
			// Tool result: return the result of a tool call back to the AI
			openaiMessages = append(openaiMessages, openai.ToolMessage(msg.Content, msg.ToolCallID))
		}
	}

	// ===== STEP 2: Convert your Tools to OpenAI format =====
	var openaiTools []openai.ChatCompletionToolParam
	for _, tool := range tools {
		openaiTools = append(openaiTools, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters:  tool.Parameters,
			},
		})
	}

	// ===== STEP 3: Call the OpenAI API =====
	params := openai.ChatCompletionNewParams{
		Messages: openaiMessages,
		Model:    "gpt-4o",
	}
	// Only add tools if there are any
	if len(openaiTools) > 0 {
		params.Tools = openaiTools
	}

	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// ===== STEP 4: Convert OpenAI response to your Response type =====
	response := &Response{
		Content: resp.Choices[0].Message.Content,
		Usage: Usage{
			InputTokens:  int(resp.Usage.PromptTokens),
			OutputTokens: int(resp.Usage.CompletionTokens),
		},
	}

	// Convert tool calls from OpenAI format to your format
	for _, tc := range resp.Choices[0].Message.ToolCalls {
		// Parse JSON arguments back into a map
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

	return response, nil
}
