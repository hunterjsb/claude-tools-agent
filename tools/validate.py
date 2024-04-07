import os
import json

def validate_tools_folder():
    tools_dir = "tools"
    
    for folder_name in os.listdir(tools_dir):
        if folder_name.startswith("__"):
            continue

        folder_path = os.path.join(tools_dir, folder_name)

        if not os.path.isdir(folder_path):
            continue

        json_file = os.path.join(folder_path, f"{folder_name}.json")
        go_file = os.path.join(folder_path, f"{folder_name}.go")

        # Check if JSON file exists and has the correct name
        if not os.path.isfile(json_file):
            print(f"Error: Missing or incorrectly named JSON file in folder '{folder_name}'")
            continue

        # Check if Go file exists and has the correct name
        if not os.path.isfile(go_file):
            print(f"Error: Missing or incorrectly named Go file in folder '{folder_name}'")
            continue

        # Check if JSON file has the correct top-level key "name"
        try:
            with open(json_file, "r") as file:
                data = json.load(file)
                if "name" not in data or data["name"] != folder_name:
                    print(f"Error: JSON file in folder '{folder_name}' does not have the correct 'name' key")
                    continue
        except json.JSONDecodeError:
            print(f"Error: Invalid JSON format in file '{json_file}'")
            continue

        # Check if Go file has the correct function name
        try:
            with open(go_file, "r") as file:
                content = file.read()
                function_name = f"func {folder_name.upper()}("
                if function_name not in content:
                    print(f"Error: Go file in folder '{folder_name}' does not have the correct function name")
                    continue
        except UnicodeDecodeError:
            print(f"Error: Unable to read Go file '{go_file}'")
            continue

        print(f"Folder '{folder_name}' passed validation")

    print("Validation complete")

# Run the validation
validate_tools_folder()