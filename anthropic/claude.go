package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hunterjsb/super-claude/config"
)

// # CLAUDE API TYPES
// - String literals for api-specific values
// - Structs for interacting with the Messages API
// - Methods for interacting with the Messages API
const MESSAGES_URL = "https://api.anthropic.com/v1/messages"

type (
	MessageRole  string
	Model        string
	StopReason   string
	ResponseType string
)

const (
	User, Assistant                        MessageRole  = "user", "assistant"
	Opus, Sonnet, Haiku                    Model        = "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"
	EndTurn, MaxTokens, StopSequence       StopReason   = "end_turn", "max_tokens", "stop_sequence"
	Text, ToolUse, MessageResp, ToolResult ResponseType = "text", "tool_use", "message", "tool_result"
)

type Message struct {
	Role    MessageRole       `json:"role"`
	Content []ResponseMessage `json:"content"`
}

type Request struct {
	Model     Model        `json:"model"`
	Messages  Conversation `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Tools     []Tool       `json:"tools,omitempty"`
}

type ResponseMessage struct {
	Type ResponseType `json:"type"`

	// text response
	Text string `json:"text,omitempty"`

	// tool_use response
	Id    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`

	// tool_response user response
	ToolUseId string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

type Response struct {
	ID           string            `json:"id"`
	Type         ResponseType      `json:"type"`
	Role         MessageRole       `json:"role"`
	Content      []ResponseMessage `json:"content"`
	Model        Model             `json:"model"`
	StopReason   StopReason        `json:"stop_reason"`
	StopSequence string            `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (r *Request) Post() (*Response, error) {
	// Marshal the JSON body
	jsonRequest, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	// Instantiate the http request
	req, err := http.NewRequest("POST", MESSAGES_URL, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.Cfg.AnthropicApiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "tools-2024-04-04")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API request failed with status code: %d, failed to read response body: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API request failed with status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	// Decode the JSON response
	var respData Response
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &respData, nil
}
