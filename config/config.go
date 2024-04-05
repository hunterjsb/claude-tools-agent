package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Cfg *Config

// # CONFIGURATION
// Config struct to type and load environment variables, and supporting methods
type Config struct {
	requireDotEnv   bool
	AnthropicApiKey string
}

func New(requireDotEnv bool) *Config {
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
