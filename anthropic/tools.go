package anthropic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

// # TOOLS
// Tools that Claude can use to take actions on the user's behalf
// They are specified in the `tools` directory as JSON files
// The name of each tool is mapped to a function
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"input_schema"`
}

type useTool func(map[string]any) Content

var ToolMap = map[string]useTool{}

type inputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Requires   []string               `json:"requires"`
}

func LoadToolFromJSONFile(filename string) (*Tool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %v", err)
	}

	var toolJSON Tool
	err = json.Unmarshal(data, &toolJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	tool := &Tool{
		Name:        toolJSON.Name,
		Description: toolJSON.Description,
		InputSchema: toolJSON.InputSchema,
	}

	return tool, nil
}

func LoadToolsFromDirectory(dir string) ([]Tool, error) {
	toolJSONs := make([]Tool, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == "tools" {
			return nil
		}

		if info.IsDir() {
			toolName := filepath.Base(path)
			toolJSONPath := filepath.Join(path, toolName+".json")
			toolGoPath := filepath.Join(path, toolName+".so")

			// Load the tool JSON file
			toolJSON, err := LoadToolFromJSONFile(toolJSONPath)
			if err != nil {
				return fmt.Errorf("failed to load tool JSON from file '%s': %v", toolJSONPath, err)
			}
			toolJSONs = append(toolJSONs, *toolJSON)

			// Load the tool's Go plugin
			plug, err := plugin.Open(toolGoPath)
			if err != nil {
				return fmt.Errorf("failed to load tool plugin from file '%s': %v", toolGoPath, err)
			}

			// Look up the UseTool function in the plugin
			useToolFunc, err := plug.Lookup(strings.ToUpper(toolName))
			if err != nil {
				return fmt.Errorf("failed to find %s function in plugin '%s': %v", toolName, toolGoPath, err)
			}

			// Assert that the UseTool function has the correct type
			useTool, ok := useToolFunc.(func(map[string]any) Content)
			if !ok {
				return fmt.Errorf("%s function in plugin '%s' has incorrect type", toolName, toolGoPath)
			}

			// Add the tool to the Tools map
			ToolMap[toolName] = useTool
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory '%s': %v", dir, err)
	}

	return toolJSONs, nil
}
