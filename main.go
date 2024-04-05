package main

import (
	"bufio"
	"log"
	"os"

	"github.com/hunterjsb/super-claude/anthropic"
	"github.com/hunterjsb/super-claude/config"
)

func main() {
	// Load config and env vars
	config.Cfg = config.New(true)
	config.Cfg.Load()

	// Get tools
	tools, err := anthropic.LoadToolsFromDirectory("tools")
	if err != nil {
		log.Fatal("FATAL: Error loading tool from JSON file.", err)
	}

	// Start the conversation
	conversation := make(anthropic.Conversation, 0)
	scanner := bufio.NewScanner(os.Stdin)
	conversation.Converse(scanner, &tools)
}
