#!/bin/bash

# Check if direnv is installed
if ! command -v direnv &> /dev/null; then
    echo "Error: direnv is not installed. Please install direnv and try again."
    exit 1
fi

# Check if env.template exists
if [ ! -f env.template ]; then
    echo "Error: env.template file not found!"
    exit 1
fi

# Temporary file for new .envrc content
temp_envrc="temp_envrc"

# Read env.template and prepend 'export ' to each line, then write to temp file
> "$temp_envrc"  # Clear temp file if it exists
while read -r line; do
    echo "export $line" >> "$temp_envrc"
done < env.template

# Check if .envrc exists, create if not
if [ ! -f .envrc ]; then
    touch .envrc
fi

# Show differences
echo "Checking for differences between .envrc and env.template..."
if diff -q .envrc "$temp_envrc" &> /dev/null; then
    # No differences
    echo "No differences found. No update needed."
    rm "$temp_envrc"
    exit 0
else
    # Show differences
    echo "Differences found:"
    diff .envrc "$temp_envrc"
fi

# If user wants to update .envrc with new content
read -p "Do you want to update .envrc with these changes? (y/n): " answer
if [[ $answer = [Yy]* ]]; then
    cp "$temp_envrc" .envrc
    echo ".envrc file updated successfully."
else
    echo "No changes made to .envrc."
fi

# Clean up
rm "$temp_envrc"
