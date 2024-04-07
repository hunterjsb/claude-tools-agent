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
            print(f"\033[91mError: Missing or incorrectly named JSON file in folder '{folder_name}' ❌\033[0m")
            continue

        # Check if Go file exists and has the correct name
        if not os.path.isfile(go_file):
            print(f"\033[91mError: Missing or incorrectly named Go file in folder '{folder_name}' ❌\033[0m")
            continue

        # Check if JSON file has the correct top-level key "name"
        try:
            with open(json_file, "r") as file:
                data = json.load(file)
                if "name" not in data or data["name"] != folder_name:
                    print(f"\033[91mError: JSON file in folder '{folder_name}' does not have the correct 'name' key ❌\033[0m")
                    continue
        except json.JSONDecodeError:
            print(f"\033[91mError: Invalid JSON format in file '{json_file}' ❌\033[0m")
            continue

        # Check if Go file has the correct function name
        try:
            with open(go_file, "r") as file:
                content = file.read()
                function_name = f"func {folder_name.upper()}("
                if function_name not in content:
                    print(f"\033[91mError: Go file in folder '{folder_name}' does not have the correct function name ❌\033[0m")
                    continue
        except UnicodeDecodeError:
            print(f"\033[91mError: Unable to read Go file '{go_file}' ❌\033[0m")
            continue

        print(f"\033[92mFolder '{folder_name}' passed validation ✅\033[0m")

    print("\033[96mValidation complete 🎉\033[0m")

# Run the validation
validate_tools_folder()