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

func (convo *Conversation) appendMsg(m Message) { // append Message to Conversation receiver
	*convo = append(*convo, m)
}

func (convo *Conversation) Converse(scanner *bufio.Scanner, t *[]Tool) {
	for {
		// Get user input (or quit)
		userInput := handleUserInput(scanner)
		if userInput == "" {
			break
		}

		// Converse
		content := makeTextContent(userInput)
		*convo = append(*convo, Message{Role: User, Content: content})
		req := &Request{Model: Opus, Messages: *convo, MaxTokens: 2048, System: SYS_PROMPT, Tools: *t}
		convo.talk(req)
	}
}

func (convo *Conversation) talk(req *Request) {
	resp, err := req.Post()
	if err != nil {
		utils.Cprintln("red", "Error making request: "+err.Error())
		return
	}

	for _, cont := range resp.Content {
		if cont.Type == MessageResp || cont.Type == Text {
			utils.Cprintln("white", "\nClaude: \n", cont.Text)
			content := make([]Content, 1)
			content[0] = cont
			convo.appendMsg(Message{Role: Assistant, Content: content})
		} else if cont.Type == ToolUse {
			convo.useTool(cont)
			req.Messages = *convo
			convo.talk(req) // Recursively call talk to handle the next step
		} else {
			utils.Cprintln("red", "Error: Unknown response type", cont.Type)
			return
		}
	}
}

func (convo *Conversation) useTool(input Content) {
	utils.Cprintln("blue", "Claude wants to use tool:", input.Name, input.Input)
	toolResp := ToolMap[input.Name](input.Input)
	currentToolUId = input.Id
	(*convo)[len(*convo)-1].Content = append((*convo)[len(*convo)-1].Content, input) // append to message content instead of conversation
	utils.Cprintln("gray", "Used tool", input.Name, "and got response", toolResp)
	convo.appendMsg(Message{Role: User, Content: makeToolResponseContent(&toolResp)})
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
