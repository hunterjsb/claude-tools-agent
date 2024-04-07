package anthropic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hunterjsb/super-claude/utils"
)

// # CONVERSATION
// Functions and logic for managing the flow of conversation with Claude
const SYS_PROMPT = `
	You are Super Claude, an AI assistant designed to help employees and developers work with Super-Sod's backend microservices. 
	We will start off by working with the 'go-postal' REST API. Use the tools provided to fulfil user requests.
	Do your best to infer user intent and take actions on their behalf.
	
	Here are some of your current directives:
	- If you need to take multiple actions, make one tool_use request, wait for the tool_results, then take the next.
	- Use common sense. Users may not be used to typing out length and precise prompts; Do your best to put together a sequence of actions to meet their needs.
	- Ask for clarification when needed.
	- If the user asks you to modify a resource, and you do so successfully, ask the user if they would like you to get the resource to confirm the modification.
	- Give brief responses - we are in dev mode and many conversations are for testing purposes.
`

const (
	claudeResponseColor = "vintage_white"
	claudeThoughtsColor = "indigo"
	toolRequestColor    = "pastel_gray"
	toolResponseColor   = "black"
	userColor           = "vintage_lime"
	claudeColor         = "pastel_pink"
)

var currentToolUId string

type Conversation []Message

func (convo *Conversation) Converse(scanner *bufio.Scanner, t *[]Tool) {
	for {
		// Get user input (or quit)
		userInput := handleUserInput(scanner)
		if userInput == "" {
			// Write conversation to JSON file on exit
			err := writeConvoToFile(*convo)
			if err != nil {
				utils.Cprintln("red", "Error writing conversation to file: "+err.Error())
			}
			break
		}

		// Converse
		content := makeTextContent(userInput)
		*convo = append(*convo, Message{Role: User, Content: content})
		req := &Request{Model: Opus, Messages: *convo, MaxTokens: 2048, System: SYS_PROMPT, Tools: *t}
		convo.talk(req)
	}
}

func makeTextContent(s string) []Content {
	content := make([]Content, 1)
	content[0] = Content{Type: Text, Text: s}
	return content
}

func (convo *Conversation) talk(req *Request) {
	resp, err := req.Post()
	// utils.Cprintln("magenta", *convo)
	if err != nil {
		utils.Cprintln("red", "Error making request: "+err.Error())
		return
	}

	for _, cont := range resp.Content {
		if cont.Type == MessageResp || cont.Type == Text {
			thoughts, message := parseThoughts(cont.Text)
			if thoughts != "" {
				utils.Cprintln(claudeThoughtsColor, "\n*Thinking* ", thoughts, "\n")
			}
			if message != "" {
				utils.Cprintln(claudeColor, "Claude:")
				utils.Cprintln(claudeResponseColor, message, "\n")
			}
			convo.appendMsg(Message{Role: Assistant, Content: wrapContent(&cont)})
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

func (convo *Conversation) appendMsg(m Message) { // append Message to Conversation receiver
	*convo = append(*convo, m)
}

func wrapContent(cont *Content) []Content {
	content := make([]Content, 1)
	content[0] = *cont
	return content
}

func (convo *Conversation) useTool(input Content) {
	utils.Cprintln(toolRequestColor, "Claude wants to use tool:", input.Name, input.Input)
	toolResp := ToolMap[input.Name](input.Input)
	currentToolUId = input.Id
	if (*convo)[len(*convo)-1].Role == Assistant {
		(*convo)[len(*convo)-1].Content = append((*convo)[len(*convo)-1].Content, input) // append to message content instead of conversation
	} else {
		convo.appendMsg(Message{Role: Assistant, Content: makeToolUseContent(&input)})
	}
	utils.Cprintln(toolResponseColor, "Used tool", input.Name, "and got response", toolResp.Content)
	convo.appendMsg(Message{Role: User, Content: makeToolResponseContent(&toolResp)})
}

func makeToolResponseContent(cont *Content) []Content {
	content := make([]Content, 1)
	content[0] = Content{Type: ToolResult, ToolUseId: currentToolUId, Content: cont.Content}
	return content
}

func makeToolUseContent(cont *Content) []Content {
	content := make([]Content, 1)
	content[0] = *cont
	return content
}

func handleUserInput(scanner *bufio.Scanner) string {
	fmt.Print(utils.Csprintf(userColor, "%s: ", "You"))
	if !scanner.Scan() {
		return ""
	}
	input := scanner.Text()
	if strings.ToLower(input) == "exit" {
		return ""
	}
	return input
}

func parseThoughts(input string) (string, string) {
	// Regular expression pattern to match the content between <thinking> tags
	pattern := `(?s)<thinking>(.*?)</thinking>`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	// Find the match in the input string
	match := re.FindStringSubmatch(input)

	// Extract the thoughts and result
	thoughts := ""
	result := ""
	if len(match) > 1 {
		thoughts = strings.TrimSpace(match[1])
		result = strings.TrimSpace(strings.Replace(input, match[0], "", 1))
	}

	if result == "" && thoughts == "" { // Fallback to original string
		result = input
	}

	return thoughts, result
}

func writeConvoToFile(convo Conversation) error {
	filename := "conversation.json"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(convo)
	if err != nil {
		return err
	}

	utils.Cprintln("green", "Conversation written to", filename)
	return nil
}
