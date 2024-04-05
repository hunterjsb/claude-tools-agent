package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

var cfg *Config

func main() {
	// Load config and env vars
	cfg = NewConfig(true)
	cfg.Load()

	// Get tools
	tools, err := LoadToolsFromDirectory("tools")
	if err != nil {
		log.Fatal("FATAL: Error loading tool from JSON file.")
	}

	// Start the conversation
	conversation := make(Conversation, 0)
	scanner := bufio.NewScanner(os.Stdin)
	conversation.Converse(scanner, &tools)
}

// # CONVERSATION
// Functions and logic for managing the flow of conversation with Claude
const SYS_PROMPT = `
	You are Super Claude, an AI assistant designed to help employees and developers work with Super-Sod's backend microservices. 
	We will start off by working with the 'go-postal' REST API. Use the tools provided to fulfil user requests.
	Do your best to infer user intent and take actions on their behalf.

	Give brief responses - we are in dev mode and many conversations are for testing purposes.
`

type Conversation []Message

func (c *Conversation) AppendResponse(msg ResponseMessage) {
	if msg.Type == text {
		newMsg := Message{Role: Assistant, Content: msg.Text}
		*c = append(*c, newMsg)
	}
}

func (c *Conversation) Converse(scanner *bufio.Scanner, tools *[]Tool) {
	for {
		input := handleUserInput(scanner)

		// Converse
		*c = append(*c, Message{Role: User, Content: input})
		req := &Request{Model: Opus, Messages: *c, MaxTokens: 2048, System: SYS_PROMPT, Tools: *tools}
		resp, err := req.Post()
		if err != nil {
			fmt.Println("Error making request: " + err.Error())
		}
		for _, msg := range resp.Content {
			if msg.Type == message || msg.Type == text {
				fmt.Printf("Claude: %s)\n", msg.Text)
				c.AppendResponse(msg)
			} else if msg.Type == toolUse {
				fmt.Println("RESPONSE TYPE TOOL USE IS NOT YET IMPLEMENTED")
			} else {
				fmt.Println("UNKNOWN RESPONSE TYPE", msg.Type)
			}
		}

	}
}

func handleUserInput(scanner *bufio.Scanner) string {
	// Get user input
	fmt.Print("You: ")
	if !scanner.Scan() {
		return ""
	}
	input := scanner.Text()
	if strings.ToLower(input) == "exit" {
		return ""
	}
	return input
}

// # TOOLS
// Tools that Claude can use to take actions on the user's behalf
// They are specified in the `tools` directory as JSON files
// The name of each tool is mapped to a function
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type UseTool func(interface{})

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Requires   []string               `json:"requires"`
}

func LoadToolFromJSONFile(filename string) (*Tool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %v", err)
	}

	var toolJSON Tool
	err = json.Unmarshal(data, &toolJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	tool := &Tool{
		Name:        toolJSON.Name,
		Description: toolJSON.Description,
		InputSchema: toolJSON.InputSchema,
	}

	return tool, nil
}

func LoadToolsFromDirectory(dir string) ([]Tool, error) {
	var tools []Tool

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".json" {
			tool, err := LoadToolFromJSONFile(path)
			if err != nil {
				return fmt.Errorf("failed to load tool from file '%s': %v", path, err)
			}
			tools = append(tools, *tool)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory '%s': %v", dir, err)
	}

	return tools, nil
}

// # CLAUDE API TYPES
// - String literals for api-specific values
// - Structs for interacting with the Messages API
// - Methods for interacting with the Messages API
const MESSAGES_URL = "https://api.anthropic.com/v1/messages"

type (
	role         string
	model        string
	stopReason   string
	responseType string
)

const (
	User, Assistant                    role         = "user", "assistant"
	Opus, Sonnet, Haiku                model        = "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"
	EndTurn, MaxTokens, StopSequence   stopReason   = "end_turn", "max_tokens", "stop_sequence"
	text, toolUse, message, toolResult responseType = "text", "tool_use", "message", "tool_result"
)

type Message struct {
	Role    role   `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     model        `json:"model"`
	Messages  Conversation `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Tools     []Tool       `json:"tools,omitempty"`
}

type ResponseMessage struct {
	Type responseType `json:"type"`

	// text response
	Text string `json:"text"`

	// tool_use response
	Id    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

type Response struct {
	ID           string            `json:"id"`
	Type         responseType      `json:"type"`
	Role         role              `json:"role"`
	Content      []ResponseMessage `json:"content"`
	Model        model             `json:"model"`
	StopReason   stopReason        `json:"stop_reason"`
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
	req.Header.Set("x-api-key", cfg.AnthropicApiKey)
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

// # CONFIGURATION
// Config struct to type and load environment variables, and supporting methods
type Config struct {
	requireDotEnv   bool
	AnthropicApiKey string
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

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: could not find ANTHROPIC_API_KEY")
	}

	c.AnthropicApiKey = apiKey
}
