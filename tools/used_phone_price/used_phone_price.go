package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hunterjsb/super-claude/anthropic"
)

func USED_PHONE_PRICE(params map[string]any) anthropic.Content {
	u, err := url.Parse("http://localhost:5000/api/iphone-used/")
	if err != nil {
		return newToolResult("ERROR Something has seriously gone wrong")
	}

	model, ok := params["phone_model"].(string)
	if !ok {
		return newToolResult("ERROR must provide phone model")
	}
	u.JoinPath(model)

	q := u.Query()
	storage := params["storage"].(string)
	unlocked := params["unlocked"].(string)
	q.Set("storage", storage)
	q.Add("unlockded", unlocked)

	resp, err := http.Get(u.String())
	if err != nil {
		errMsg := "ERROR on http.Get: " + err.Error()
		return newToolResult(errMsg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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