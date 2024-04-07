# super-claude
### An autonomous agent based on Anthropic's `tools` beta and Claude-3

## Usage
`super-claude` is currently used via command-line. It allows you to chat with Claude and instruct it to use tools on your behalf. Tools are golang plugins in the `tools/` directory.
- Run super-claude from source: `$ go run main.go` 
- Build and run: `$ build.sh`
- Extract and run: `$ tar -xzf super-claude.tar.gz && ./super-claude`

## Tools
super-claude can use the tools in the `tools/` directory, which are written in Go and compiled as plugins. A tool has two components:
- A **JSON schema** which defines the name, description, and parameters of a tool
- A **Go plugin** which has the following signature: 
    ```Go
    func TOOL_NAME(map[param]any) anthropic.Content
    ```
A Tool needs to be structured as follows:

```
.
├── tools
│   ├── build.sh
│   ├── tool_name
│   │   ├── tool_name.go
│   │   ├── tool_name.json
|   |   └── tool_name.so
│   └── validate.py
```

The `"name"` top-level attribute in tool_name.json should also be tool_name. 
#### Validating Tools:
Run `tools/validate.py` to make sure your files and functions are named correctly.
![validate](https://i.imgur.com/JTJT8DK.gif)

#### Compiling Tools:
`tools/build.sh` will iterate over `tools/` and compile each Go file as a plugin.
![compile tools](https://i.imgur.com/WLym53o.gif)


#### Using compiled tools
Claude can deal with chain tool usages and deal with errors as seen here
![use tools](https://i.imgur.com/UxDWiyr.gif)