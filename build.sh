#!/bin/bash

# Run validate.py
echo "Running validate.py..."
python3 tools/validate.py
if [ $? -ne 0 ]; then
    echo "Validation failed. Exiting."
    exit 1
fi

# Run build.sh
echo "Running build.sh..."
./tools/build.sh
if [ $? -ne 0 ]; then
    echo "Build failed. Exiting."
    exit 1
fi

# Build Go plugins
echo "Building Go plugins..."
for tool_dir in tools/*/; do
    if [ -d "$tool_dir" ]; then
        tool_name=$(basename "$tool_dir")
        go_file="$tool_dir/$tool_name.go"
        so_file="$tool_dir/$tool_name.so"

        if [ -f "$go_file" ]; then
            echo "Building plugin: $tool_name"
            go build -buildmode=plugin -o "$so_file" "$go_file"
            if [ $? -ne 0 ]; then
                echo "Failed to build plugin: $tool_name"
                exit 1
            fi
        fi
    fi
done

# Build the main Go project
echo "Building the main Go project..."
go build -o super-claude
if [ $? -ne 0 ]; then
    echo "Go build failed. Exiting."
    exit 1
fi

# Create a temporary directory for packaging
mkdir -p release

# Copy the main executable to the release directory
cp super-claude release/

# Copy the tools directory to the release directory
cp -R tools release/

# Create the tarball
tar -czf super-claude.tar.gz -C release .

# Remove the temporary release directory
rm -rf release

echo "Tarball created: super-claude.tar.gz"

# Run the main Go project
echo "Running the main Go project..."
./super-claude