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
	// Could append the Content AS a Message to the Conversation
	// OR Content to a Message depending on the Content type
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
		utils.Cprintln("gray", "Ignoring message of type", cont.Type)
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
		utils.Cprintln("red", "Error making request: "+err.Error())
		return
	}

	for _, cont := range resp.Content {
		if cont.Type == MessageResp || cont.Type == Text {
			utils.Cprintln("white", "\nClaude: \n", cont.Text)
			c.appendContent(cont)
		} else if cont.Type == ToolUse {
			c.useTool(cont)

			// Send a new request to Claude asking for the next step or a summary
			req.Messages = *c
			newResp, err := req.Post()
			if err != nil {
				utils.Cprintln("red", "Error making request: "+err.Error())
				return
			}

			// Process the new response from Claude
			for _, newMsg := range newResp.Content {
				if newMsg.Type == MessageResp || newMsg.Type == Text {
					utils.Cprintln("white", "\nClaude: ", newMsg.Text)
					c.appendContent(newMsg)
				} else {
					utils.Cprintln("red", "Error: Cannot chain actions!", newMsg.Type)
					return
				}
			}
		} else {
			utils.Cprintln("red", "Error: Unknown response type", cont.Type)
			return
		}
	}
}

func (c *Conversation) useTool(input Content) {
	utils.Cprintln("blue", "Claude wants to use tool:", input.Name, input.Input)
	toolResp := ToolMap[input.Name](input.Input)
	c.appendContent(input)
	utils.Cprintln("gray", "Used tool", input.Name, "and got response", toolResp)
	c.appendMsg(Message{Role: User, Content: makeToolResponseContent(&toolResp)})
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
