package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ksysoev/pdd-action/pkg/core"
	"github.com/ksysoev/pdd-action/pkg/github"
)

func main() {
	// Read token from environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN environment variable is required")
		os.Exit(1)
	}

	repoFullName := os.Getenv("GITHUB_REPOSITORY")
	if repoFullName == "" {
		repoFullName = "ksysoev/pdd-action" // Default for testing
	}

	// Initialize config with test branch
	config := core.Config{
		GitHubToken:      token,
		BranchName:       "test-pdd-action",
		IssueTitlePrefix: "[TEST PDD]",
	}

	// Initialize GitHub client
	client := github.NewClient(token, repoFullName, config)

	// Create test TODO comments
	comments := []core.TodoComment{
		{
			FilePath:    "test-todo-comments.go",
			LineNumber:  10,
			Title:       "Test issue creation",
			Description: []string{"This is a test description", "For checking if issue creation works"},
			Labels:      []string{"test", "pdd"},
		},
	}

	// Create issues from comments
	ctx := context.Background()
	processedComments, err := client.CreateIssuesFromComments(ctx, comments)
	if err != nil {
		fmt.Printf("Error creating issues: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created %d issues\n", len(processedComments))
	for i, comment := range processedComments {
		fmt.Printf("Issue %d: %s\n", i+1, comment.IssueURL)
	}
}