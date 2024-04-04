package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var cfg *Config

func main() {
	cfg = NewConfig(true)
	cfg.Load()

	conversation := make([]Message, 1)
	conversation[0] = Message{Role: User, Content: "Hello, world!"}

	req := &Request{Model: Opus, Messages: conversation, MaxTokens: 2048}
	resp, err := req.Post()
	if err != nil {
		fmt.Println("Error making request: " + err.Error())
	} else {
		fmt.Println("Claude responded with:", resp)
	}
}

// # CLAUDE API TYPES
// - String literals for api-specific values
// - Structs for interacting with the Messages API
// - Methods for interacting with the Messages API
const MESSAGES_URL = "https://api.anthropic.com/v1/messages"

type (
	role       string
	model      string
	stopReason string
)

const (
	User, Assistant                  role       = "user", "assistant"
	Opus, Sonnet, Haiku              model      = "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"
	EndTurn, MaxTokens, StopSequence stopReason = "end_turn", "max_tokens", "stop_sequence"
)

type Message struct {
	Role    role   `json:"role"`    // required
	Content string `json:"content"` //required
	System  string `json:"system,omitempty"`
}

type Request struct {
	Model     model     `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Response struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    role   `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        model      `json:"model"`
	StopReason   stopReason `json:"stop_reason"`
	StopSequence string     `json:"stop_sequence"`
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
	req.Header.Set("x-api-key", cfg.ClaudeApiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

		// Convert the response body to a string
		bodyString := string(body)

		return nil, fmt.Errorf("API request failed with status code: %d, response body: %s", resp.StatusCode, bodyString)
	}

	// Decode the JSON response
	var respData Response
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &respData, nil
}

// # CONFIGURATION
// Config struct to type and load environment variables, and supporting methods
type Config struct {
	requireDotEnv bool
	ClaudeApiKey  string
}

func NewConfig(requireDotEnv bool) *Config {
	return &Config{requireDotEnv: requireDotEnv}
}

func (c *Config) Load() {
	err := godotenv.Load()
	if err != nil {
		if c.requireDotEnv {
			log.Fatal("FATAL: Could not load .env")
		} else {
			fmt.Println("Could not load .env, continuing...")
		}
	}

	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: could not find CLAUDE_API_KEY")
	}

	c.ClaudeApiKey = apiKey
}
