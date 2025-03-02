#!/bin/bash

# Check if git command is available
if ! command -v git &> /dev/null; then
    echo "git command not found. Please install git."
    exit 1
fi

# Check if go command is available
if ! command -v go &> /dev/null; then
    echo "go command not found. Please install Go language."
    exit 1
fi

# Check Go language version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_VERSION_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_VERSION_MINOR=$(echo $GO_VERSION | cut -d. -f2)

if [ "$GO_VERSION_MAJOR" -lt 1 ] || [ "$GO_VERSION_MAJOR" -eq 1 -a "$GO_VERSION_MINOR" -lt 23 ]; then
    echo "Go language version is too old. Go 1.23.0 or higher is required."
    echo "Current version: $GO_VERSION"
    exit 1
fi

# Build the project
echo "Building cocommit..."
go build -o bin/git-cocommit cmd/cocommit/main.go

# Get path
INSTALL_PATH="$(git --exec-path)"

# Copy the executable
echo "Installing cocommit command to ${INSTALL_PATH}..."
cp bin/git-cocommit "${INSTALL_PATH}/git-cocommit"

echo "Installation completed!"
echo "Usage: git cocommit -m \"Commit message\"" 