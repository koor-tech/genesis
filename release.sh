#!/bin/bash

# Check if a tag was passed as an argument
if [ -n "$1" ]; then
    # Use the provided tag
    new_tag=$1
    echo "Tag provided: $new_tag"
else
    # Fetch the latest tag from Git
    last_tag=$(git describe --tags `git rev-list --tags --max-count=1`)

    # If no previous tags are found, set a default initial version
    if [ -z "$last_tag" ]; then
        echo "No previous tags found. Setting initial version."
        new_tag="0.0.1"
    else
        echo "Last tag found: $last_tag"
        # Automatically increment the version
        new_tag=$(echo $last_tag | awk -F. '{printf("%d.%d.%d", $1, $2, $3+1)}')
    fi

    echo "New tag: $new_tag"

    # Ask for confirmation
    read -p "Is this the correct tag? (y/n) " -n 1 -r
    echo    # Move to a new line
    if [[ ! $REPLY =~ ^[Yy]$ ]]
    then
        echo "Aborting."
        exit 1
    fi
fi

# Create the new tag
git tag $new_tag

# Push the new tag
git push origin $new_tag

echo "Tag $new_tag created and pushed successfully."
