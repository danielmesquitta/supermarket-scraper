#!/bin/bash

packages=(
    "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    "github.com/segmentio/golines@latest"
    "github.com/google/wire/cmd/wire@latest"
    "github.com/danielmesquitta/prisma-go-tools@latest"
)

echo "Installing and updating Go packages..."

for package in "${packages[@]}"; do
    echo "$package..."
    go install "$package"
done

echo "All packages have been successfully installed and updated."
