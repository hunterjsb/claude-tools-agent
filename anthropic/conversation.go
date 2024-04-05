package anthropic

import (
	"bufio"
	"fmt"
	"strings"
)

// # CONVERSATION
// Functions and logic for managing the flow of conversation with Claude
const SYS_PROMPT = `
	You are Super Claude, an AI assistant designed to help employees and developers work with Super-Sod's backend microservices. 
	We will start off by working with the 'go-postal' REST API. Use the tools provided to fulfil user requests.
	Do your best to infer user intent and take actions on their behalf.

	Give brief responses - we are in dev mode and many conversations are for testing purposes.
`

var currentToolUId string

type Conversation []Message

func (c *Conversation) AppendResponse(msg ResponseMessage, data *ResponseMessage) {
	content := make([]ResponseMessage, 1)
	if msg.Type == Text {
		content[0] = msg
		newMsg := Message{Role: Assistant, Content: content}
		*c = append(*c, newMsg)
	} else if msg.Type == ToolResult {
		msg.ToolUseId = currentToolUId
		content[0] = msg
		newMsg := Message{Role: User, Content: content}
		*c = append(*c, newMsg)
	} else if msg.Type == ToolUse {
		fmt.Println("tool_use not logged for ID", msg.Id)
		currentToolUId = msg.Id
		(*c)[len(*c)-1].Content = append((*c)[len(*c)-1].Content, msg)
	} else {
		fmt.Println("Ignoring message of type", msg.Type)
	}
}

func (c *Conversation) Converse(scanner *bufio.Scanner, t *[]Tool) {
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
		return
	}

	for _, msg := range resp.Content {
		if msg.Type == MessageResp || msg.Type == Text {
			fmt.Printf("\nClaude: %s)\n", msg.Text)
			c.AppendResponse(msg, nil)
		} else if msg.Type == ToolUse {
			fmt.Println("\nClaude wants to use tool:", msg.Name, msg.Input)
			toolResp, err := ToolMap[msg.Name](msg.Input)
			if err != nil {
				fmt.Println("ERROR using tool", err)
				return
			}

			c.AppendResponse(msg, toolResp)
			fmt.Println("Used tool", msg.Name, "and got response", toolResp)
			toolResultMsg := Message{Role: User, Content: makeToolResponseContent(toolResp)}
			*c = append(*c, toolResultMsg)
			fmt.Println("TOOL RESPONSE:::", *c)

			// Send a new request to Claude asking for the next step or a summary
			newReq := &Request{Model: Opus, Messages: *c, MaxTokens: 2048, System: SYS_PROMPT, Tools: req.Tools}
			newResp, err := newReq.Post()
			if err != nil {
				fmt.Println("Error making request: " + err.Error())
				return
			}

			// Process the new response from Claude
			for _, newMsg := range newResp.Content {
				if newMsg.Type == MessageResp || newMsg.Type == Text {
					fmt.Printf("\nClaude: %s)\n", newMsg.Text)
					c.AppendResponse(newMsg, nil)
				} else {
					fmt.Println("Error: Unknown response type", newMsg.Type)
					return
				}
			}
		} else {
			fmt.Println("Error: Unknown response type", msg.Type)
			return
		}
	}
}

func makeTextContent(s string) []ResponseMessage {
	content := make([]ResponseMessage, 1)
	content[0] = ResponseMessage{Type: Text, Text: s}
	return content
}

func makeToolResponseContent(data *ResponseMessage) []ResponseMessage {
	content := make([]ResponseMessage, 1)
	content[0] = ResponseMessage{Type: ToolResult, ToolUseId: currentToolUId, Content: data.Content}
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
