package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hunterjsb/super-claude/anthropic"
)

func GET_POSTAL_CODES(params map[string]any) (*anthropic.Content, error) {
	postalCode, ok := params["postal_code"]
	if !ok {
		return nil, errors.New("must provide postal_code")
	}
	url := fmt.Sprintf("http://localhost:8280/zipcodes/%s", postalCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	fmt.Println("GET_POSTAL_CODES got response", resp.StatusCode)
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API request failed with status code: %d, failed to read response body: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API request failed with status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	// Decode the JSON response
	responseContent, err := (io.ReadAll(resp.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &anthropic.Content{Type: anthropic.ToolResult, Content: string(responseContent)}, nil
}
