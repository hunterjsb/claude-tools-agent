package anthropic

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/hunterjsb/super-claude/utils"
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

func (c *Conversation) appendMsg(m Message) { // append Message to Conversation receiver
	*c = append(*c, m)
}

func (c *Conversation) appendContent(cont Content) {
	// Could append a message to the conversation OR content to a message depending on the content type
	content := make([]Content, 1)
	if cont.Type == Text {
		content[0] = cont
		c.appendMsg(Message{Role: Assistant, Content: content})
	} else if cont.Type == ToolResult {
		cont.ToolUseId = currentToolUId
		content[0] = cont
		c.appendMsg(Message{Role: User, Content: content})
	} else if cont.Type == ToolUse {
		currentToolUId = cont.Id
		(*c)[len(*c)-1].Content = append((*c)[len(*c)-1].Content, cont) // append to content instead of conversation
	} else {
		utils.CPrint("gray", "Ignoring message of type", cont.Type)
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
		c.talk(req)
	}
}

func (c *Conversation) talk(req *Request) {
	resp, err := req.Post()
	if err != nil {
		utils.CPrint("red", "Error making request: "+err.Error())
		return
	}

	for _, msg := range resp.Content {
		if msg.Type == MessageResp || msg.Type == Text {
			utils.CPrint("white", "\nClaude: \n", msg.Text)
			c.appendContent(msg)
		} else if msg.Type == ToolUse {
			utils.CPrint("blue", "Claude wants to use tool:", msg.Name, msg.Input)
			toolResp, err := ToolMap[msg.Name](msg.Input)
			if err != nil {
				utils.CPrint("red", "ERROR using tool", err)
				return
			}

			c.appendContent(msg)
			utils.CPrint("gray", "Used tool", msg.Name, "and got response", toolResp)
			toolResultMsg := Message{Role: User, Content: makeToolResponseContent(toolResp)}
			*c = append(*c, toolResultMsg)
			// fmt.Println("TOOL RESPONSE:::", *c)

			// Send a new request to Claude asking for the next step or a summary
			newReq := &Request{Model: Opus, Messages: *c, MaxTokens: 2048, System: SYS_PROMPT, Tools: req.Tools}
			newResp, err := newReq.Post()
			if err != nil {
				utils.CPrint("red", "Error making request: "+err.Error())
				return
			}

			// Process the new response from Claude
			for _, newMsg := range newResp.Content {
				if newMsg.Type == MessageResp || newMsg.Type == Text {
					utils.CPrint("white", "\nClaude: %s)\n", newMsg.Text)
					c.appendContent(newMsg)
				} else {
					utils.CPrint("red", "Error: Unknown response type", newMsg.Type)
					return
				}
			}
		} else {
			utils.CPrint("red", "Error: Unknown response type", msg.Type)
			return
		}
	}
}

func makeTextContent(s string) []Content {
	content := make([]Content, 1)
	content[0] = Content{Type: Text, Text: s}
	return content
}

func makeToolResponseContent(data *Content) []Content {
	content := make([]Content, 1)
	content[0] = Content{Type: ToolResult, ToolUseId: currentToolUId, Content: data.Content}
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
