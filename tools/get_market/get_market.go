package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hunterjsb/super-claude/anthropic"
)

func GET_MARKET(params map[string]any) anthropic.Content {
	postalCode, ok := params["market_id"]
	if !ok {
		return newToolResult("ERROR: must provide market_id")
	}

	pocoFloat, ok := postalCode.(float64) // idk why it comes over as float64
	if !ok {
		return newToolResult("ERROR: invalid market_id type")
	}

	url := fmt.Sprintf("http://localhost:8280/markets/%d", int(pocoFloat))
	resp, err := http.Get(url)
	if err != nil {
		errMsg := "ERROR on http.Get: " + err.Error()
		return newToolResult(errMsg)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := fmt.Sprintf("API request failed with status code: %d, failed to read response body: %v", resp.StatusCode, err)
			return newToolResult(errMsg)
		}
		errMsg := fmt.Sprintf("API request failed with status code: %d, response body: %s", resp.StatusCode, string(body))
		return newToolResult(errMsg)
	}

	// Decode the JSON response
	responseContent, err := (io.ReadAll(resp.Body))
	if err != nil {
		return newToolResult(fmt.Sprintf("failed to decode response: %v", err))
	}

	return newToolResult(string(responseContent))
}

func newToolResult(s string) anthropic.Content {
	return anthropic.Content{Type: anthropic.ToolResult, Content: s}
}
