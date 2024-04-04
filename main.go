package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	cfg := newConfig(true)
	cfg.load()
}

type Config struct {
	requireDotEnv bool
	claudeApiKey  string
}

func newConfig(requireDotEnv bool) *Config {
	return &Config{requireDotEnv: requireDotEnv}
}

func (c *Config) load() {
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

	c.claudeApiKey = apiKey
}
