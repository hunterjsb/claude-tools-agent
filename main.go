package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/hunterjsb/super-claude/anthropic"
	"github.com/hunterjsb/super-claude/config"
)

func main() {
	// Define command-line flags
	startServer := flag.Bool("server", false, "Start the HTTP server")
	flag.Parse()

	// Load config and env vars
	config.Cfg = config.New(true)
	config.Cfg.Load()

	// Get tools
	tools, err := anthropic.LoadToolsFromDirectory("tools")
	if err != nil {
		log.Fatal("FATAL: Error loading tool from JSON file.", err)
	}

	conversation := make(anthropic.Conversation, 0)
	if *startServer {
		// Start the HTTP server
		handler := anthropic.Handler{Tools: &tools}
		http.HandleFunc("/", handler.ConverseHttp)

		log.Println("Starting HTTP server on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	} else {
		// Start the conversation
		scanner := bufio.NewScanner(os.Stdin)
		conversation.Converse(scanner, &tools)
	}
}
