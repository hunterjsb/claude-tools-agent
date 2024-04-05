#!/bin/bash

# Set the tools directory
tools_dir="tools"

# Iterate over each subdirectory in the tools directory
for tool_dir in "$tools_dir"/*; do
  if [ -d "$tool_dir" ]; then
    # Get the tool name from the directory name
    tool_name=$(basename "$tool_dir")
    
    # Set the paths for the Go file and the output plugin file
    go_file="$tool_dir/$tool_name.go"
    plugin_file="$tool_dir/$tool_name.so"
    
    # Check if the Go file exists
    if [ -f "$go_file" ]; then
      echo "Building plugin for $tool_name..."
      
      # Build the plugin using the Go file
      go build -buildmode=plugin -o "$plugin_file" "$go_file"
      
      if [ $? -eq 0 ]; then
        echo "Plugin for $tool_name built successfully."
      else
        echo "Failed to build plugin for $tool_name."
      fi
    else
      echo "Go file not found for $tool_name. Skipping."
    fi
    
    echo "------------------------"
  fi
done