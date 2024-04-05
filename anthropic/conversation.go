package anthropic

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/hunterjsb/super-claude/tools"
)

// # CONVERSATION
// Functions and logic for managing the flow of conversation with Claude
const SYS_PROMPT = `
	You are Super Claude, an AI assistant designed to help employees and developers work with Super-Sod's backend microservices. 
	We will start off by working with the 'go-postal' REST API. Use the tools provided to fulfil user requests.
	Do your best to infer user intent and take actions on their behalf.

	Give brief responses - we are in dev mode and many conversations are for testing purposes.
`

// var currentToolUId string

type Conversation []Message

func (c *Conversation) AppendResponse(msg ResponseMessage, data *any) {
	content := make([]ResponseMessage, 1)
	if msg.Type == text {
		content[0] = msg
		newMsg := Message{Role: Assistant, Content: content}
		*c = append(*c, newMsg)
	} else if msg.Type == toolResult {
		content[0] = msg
		newMsg := Message{Role: User, Content: content}
		*c = append(*c, newMsg)
	} else if msg.Type == toolUse {
		fmt.Println("tool_use not logged")
		// (*c)[len(*c)-1].Content = (*c)[len(*c)-1].Content + fmt.Sprintf("%v", msg.Input)
	} else {
		fmt.Println("Ignoring message of type", msg.Type)
	}
}

func (c *Conversation) Converse(scanner *bufio.Scanner, t *[]tools.Tool) {
	for {
		// Get user input (or quit)
		userInput := handleUserInput(scanner)
		if userInput == "" {
			break
		}

		// Converse
		content := makeTextContent(userInput)
		*c = append(*c, Message{Role: User, Content: content})
		req := &Request{Model: Opus, Messages: *c, MaxTokens: 2048, System: SYS_PROMPT, Tools: *t}
		c.loop(req)
	}
}

func (c *Conversation) loop(req *Request) {
	resp, err := req.Post()
	if err != nil {
		fmt.Println("Error making request: " + err.Error())
	}
	for _, msg := range resp.Content {
		if msg.Type == message || msg.Type == text {
			fmt.Printf("\nClaude: %s)\n", msg.Text)
			c.AppendResponse(msg, nil)
		} else if msg.Type == toolUse {
			fmt.Println("\nClaude wants to use tool:", msg.Name, msg.Input)
			toolResp, err := tools.ToolMap[msg.Name](msg.Input)
			if err != nil {
				fmt.Println("ERROR using tool", err)
			} else {
				c.AppendResponse(msg, &toolResp)
				fmt.Println("Used tool", msg.Name, "and got response", toolResp)
				toolResultMsg := Message{Role: User, Content: makeToolResponseContent(toolResp)}
				*c = append(*c, toolResultMsg)
				fmt.Println("TOOL RESPONSE:::", *c)
				c.loop(req) // recursion
			}
		} else {
			fmt.Println("Error: Unknown response type", msg.Type)
		}
	}
}

func makeTextContent(s string) []ResponseMessage {
	content := make([]ResponseMessage, 1)
	content[0] = ResponseMessage{Type: text, Text: s}
	return content
}

func makeToolResponseContent(data any) []ResponseMessage {
	data, ok := data.(any)
	fmt.Println(data, ok)
	content := make([]ResponseMessage, 1)
	content[0] = ResponseMessage{Type: text, Text: "PLACEHOLDER"}
	return content
}

func handleUserInput(scanner *bufio.Scanner) string {
	fmt.Print("\nYou: ")
	if !scanner.Scan() {
		return ""
	}
	input := scanner.Text()
	if strings.ToLower(input) == "exit" {
		return ""
	}
	return input
}
