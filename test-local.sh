#!/bin/bash

# Set environment variables to simulate GitHub Actions environment
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Please set GITHUB_TOKEN environment variable before running this script"
    exit 1
fi

export GITHUB_REPOSITORY="ksysoev/pdd-action"
export GITHUB_EVENT_NAME="workflow_dispatch"
export GITHUB_WORKSPACE=$(pwd)
export GITHUB_REF_NAME="test-pdd-action"

echo "Running test with:"
echo "Repository: $GITHUB_REPOSITORY"
echo "Event: $GITHUB_EVENT_NAME"
echo "Workspace: $GITHUB_WORKSPACE"
echo "Branch: $GITHUB_REF_NAME"
echo "Token length: ${#GITHUB_TOKEN}"

# Run the action with required inputs
./pdd-action