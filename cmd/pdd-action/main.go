package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ksysoev/pdd-action/pkg/core"
	"github.com/ksysoev/pdd-action/pkg/github"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	// Set up action
	action := githubactions.New()
	ctx := context.Background()

	// Get action inputs - first try action inputs, then fall back to env vars
	githubToken := action.GetInput("github_token")
	if githubToken == "" {
		// Try various environment variable names that might contain the token
		githubToken = os.Getenv("PDD_GITHUB_TOKEN")
		if githubToken == "" {
			githubToken = os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				// Final fallback to catch other possible env var names
				githubToken = os.Getenv("GH_TOKEN")
				if githubToken == "" {
					action.Fatalf("github_token input is required")
				}
			}
		}
	}
	
	action.Infof("GitHub token is present with length: %d", len(githubToken))

	branchName := action.GetInput("branch_name")
	if branchName == "" {
		branchName = os.Getenv("PDD_BRANCH_NAME")
		if branchName == "" {
			branchName = "main" // Default branch name
		}
	}

	issueTitlePrefix := action.GetInput("issue_title_prefix")
	if issueTitlePrefix == "" {
		issueTitlePrefix = os.Getenv("PDD_ISSUE_PREFIX")
	}

	// Get GitHub context
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if eventName != "pull_request" && eventName != "workflow_dispatch" && eventName != "push" {
		action.Fatalf("This action only works on pull_request, workflow_dispatch, or push events, got: %s", eventName)
	}

	repoFullName := os.Getenv("GITHUB_REPOSITORY")
	if repoFullName == "" {
		action.Fatalf("GITHUB_REPOSITORY environment variable is not set")
	}

	var prNumber int
	var err error
	
	if eventName == "workflow_dispatch" || eventName == "push" {
		// In workflow_dispatch or push mode, use a dummy PR number
		prNumber = 1
		action.Infof("Running in %s mode - using dummy PR number: %d", eventName, prNumber)
	} else {
		prString := os.Getenv("GITHUB_REF")
		prNumber, err = extractPRNumber(prString)
		if err != nil {
			action.Fatalf("Failed to extract PR number: %v", err)
		}
	}

	// Get workspace path
	workspacePath := os.Getenv("GITHUB_WORKSPACE")
	if workspacePath == "" {
		action.Fatalf("GITHUB_WORKSPACE environment variable is not set")
	}

	// Initialize config
	config := core.Config{
		GitHubToken:      githubToken,
		BranchName:       branchName,
		IssueTitlePrefix: issueTitlePrefix,
	}

	// Initialize GitHub client
	client := github.NewClient(githubToken, repoFullName, config)

	// For workflow_dispatch or push, skip PR merged check
	if eventName != "workflow_dispatch" && eventName != "push" {
		// Check if PR is merged to target branch
		isMerged, err := client.IsPRMergedToTargetBranch(ctx, prNumber)
		if err != nil {
			action.Fatalf("Failed to check if PR is merged: %v", err)
		}
	
		if !isMerged {
			action.Infof("PR #%d is not merged to %s branch yet. Skipping issue creation.", prNumber, branchName)
			return
		}
	} else {
		action.Infof("Running in %s mode - skipping PR merged check", eventName)
		action.Infof("Using target branch for issues: %s", branchName)
	}

	// Scan workspace for TODO comments
	excludeDirs := []string{
		filepath.Join(workspacePath, ".git"),
		filepath.Join(workspacePath, "node_modules"),
		filepath.Join(workspacePath, "vendor"),
	}

	action.Infof("Scanning for TODO comments in workspace: %s", workspacePath)
	comments, err := core.ScanDirectory(workspacePath, excludeDirs)
	if err != nil {
		action.Fatalf("Failed to scan directory: %v", err)
	}

	action.Infof("Found %d TODO comments", len(comments))

	// Filter out already processed comments
	unprocessedComments := core.FilterUnprocessedComments(comments)
	action.Infof("Found %d unprocessed TODO comments", len(unprocessedComments))

	if len(unprocessedComments) == 0 {
		action.Infof("No unprocessed TODO comments found. Exiting.")
		return
	}

	// Create issues from unprocessed comments
	processedComments, err := client.CreateIssuesFromComments(ctx, unprocessedComments)
	if err != nil {
		action.Fatalf("Failed to create issues: %v", err)
	}

	action.Infof("Created %d issues from TODO comments", len(processedComments))

	// Get PR head branch name or use current branch for workflow_dispatch/push
	var prBranch string
	if eventName == "workflow_dispatch" || eventName == "push" {
		// Use the configured branch or fallback to GitHub ref
		prBranch = branchName
		
		// For development testing, use actual git branch if possible
		if prBranch == "main" && os.Getenv("GITHUB_REF_NAME") != "" {
			prBranch = os.Getenv("GITHUB_REF_NAME")
		}
		
		action.Infof("Using branch for %s: %s", eventName, prBranch)
	} else {
		githubClient := github.NewRawClient(githubToken)
		
		// Split repository owner and name safely
		repoParts := strings.Split(repoFullName, "/")
		owner := repoParts[0]
		repo := ""
		if len(repoParts) > 1 {
			repo = repoParts[1]
		} else {
			repo = repoFullName
			owner = os.Getenv("GITHUB_REPOSITORY_OWNER")
			if owner == "" {
				owner = "unknown"
			}
		}
		
		prDetails, _, err := githubClient.PullRequests.Get(ctx, owner, repo, prNumber)
		if err != nil {
			action.Fatalf("Failed to get PR details: %v", err)
		}
		prBranch = prDetails.GetHead().GetRef()
	}

	// Update comments in PR files
	for _, comment := range processedComments {
		err := client.UpdateCommentInFile(ctx, comment, prNumber, prBranch)
		if err != nil {
			action.Warningf("Failed to update comment in file %s: %v", comment.FilePath, err)
		} else {
			action.Infof("Updated TODO comment in %s with issue URL: %s", comment.FilePath, comment.IssueURL)
		}
	}

	action.Infof("PDD Action completed successfully")
}

// extractPRNumber extracts the PR number from the GITHUB_REF
func extractPRNumber(refString string) (int, error) {
	// GitHub Actions format: refs/pull/{number}/merge
	pullPrefix := "refs/pull/"
	if strings.HasPrefix(refString, pullPrefix) {
		numStr := strings.TrimPrefix(refString, pullPrefix)
		numStr = strings.Split(numStr, "/")[0]
		return strconv.Atoi(numStr)
	}

	// If we can't extract from GITHUB_REF, try GITHUB_REF_NAME (which might be just the number in some cases)
	refName := os.Getenv("GITHUB_REF_NAME")
	if refName != "" {
		// Try to extract number from the ref name (might be "123/merge" or just "123")
		parts := strings.Split(refName, "/")
		return strconv.Atoi(parts[0])
	}

	// Try to get it from GITHUB_EVENT_PATH
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath != "" {
		data, err := os.ReadFile(eventPath)
		if err == nil {
			// Extract PR number from event payload JSON
			// This is a simple string search, not proper JSON parsing
			prMarker := "\"number\":"
			if idx := strings.Index(string(data), prMarker); idx > 0 {
				numStart := idx + len(prMarker)
				numEnd := strings.Index(string(data[numStart:]), ",")
				if numEnd > 0 {
					numStr := strings.TrimSpace(string(data[numStart : numStart+numEnd]))
					return strconv.Atoi(numStr)
				}
			}
		}
	}

	return 0, fmt.Errorf("could not extract PR number from refs: %s", refString)
}
