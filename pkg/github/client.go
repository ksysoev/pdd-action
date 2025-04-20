package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/kirill/pdd-action/pkg/core"
	"golang.org/x/oauth2"
)

// Client handles interaction with the GitHub API
type Client struct {
	client *github.Client
	owner  string
	repo   string
	config core.Config
}

// NewClient creates a new GitHub client
func NewClient(token, repoFullName string, config core.Config) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	parts := strings.Split(repoFullName, "/")
	var owner, repo string
	
	fmt.Printf("Repository full name: %s\n", repoFullName)
	
	if len(parts) >= 2 {
		owner = parts[0]
		repo = parts[1]
	} else {
		// Fallback to environment variables if possible
		owner = os.Getenv("GITHUB_REPOSITORY_OWNER")
		if owner == "" {
			owner = "unknown"
		}
		repo = repoFullName // Use as-is if can't split
	}
	
	fmt.Printf("Repository owner: %s, repo: %s\n", owner, repo)

	return &Client{
		client: client,
		owner:  owner,
		repo:   repo,
		config: config,
	}
}

// NewRawClient creates a new raw GitHub client
func NewRawClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// CreateIssuesFromComments creates GitHub issues from TODO comments
func (c *Client) CreateIssuesFromComments(ctx context.Context, comments []core.TodoComment) ([]core.TodoComment, error) {
	var processedComments []core.TodoComment

	fmt.Printf("Creating issues for %d comments in repository %s/%s\n", len(comments), c.owner, c.repo)

	for _, comment := range comments {
		// Skip comments that already have an issue URL
		if comment.IssueURL != "" {
			continue
		}

		// Prepare issue title with optional prefix
		title := comment.Title
		if c.config.IssueTitlePrefix != "" {
			title = fmt.Sprintf("%s %s", c.config.IssueTitlePrefix, title)
		}

		// Prepare issue body
		body := fmt.Sprintf("Created from TODO comment in `%s` (line %d):\n\n", comment.FilePath, comment.LineNumber)
		body += strings.Join(comment.Description, "\n")

		fmt.Printf("Creating issue with title: %s\n", title)
		fmt.Printf("Labels: %v\n", comment.Labels)

		// Create the issue
		issue, resp, err := c.client.Issues.Create(ctx, c.owner, c.repo, &github.IssueRequest{
			Title:  &title,
			Body:   &body,
			Labels: &comment.Labels,
		})
		if err != nil {
			fmt.Printf("Error creating issue: %v\n", err)
			if resp != nil {
				fmt.Printf("Response status: %s\n", resp.Status)
			}
			return processedComments, fmt.Errorf("failed to create issue for comment in %s (line %d): %w",
				comment.FilePath, comment.LineNumber, err)
		}

		// Update the comment with the issue URL
		comment.IssueURL = issue.GetHTMLURL()
		fmt.Printf("Created issue: %s\n", comment.IssueURL)
		processedComments = append(processedComments, comment)
	}

	return processedComments, nil
}

// IsPRMergedToTargetBranch checks if a PR is merged to the target branch
func (c *Client) IsPRMergedToTargetBranch(ctx context.Context, prNumber int) (bool, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, prNumber)
	if err != nil {
		return false, fmt.Errorf("failed to get PR #%d: %w", prNumber, err)
	}

	// Check if PR is merged and merges to the configured target branch
	return pr.GetMerged() && pr.GetBase().GetRef() == c.config.BranchName, nil
}

// UpdateCommentInFile updates the TODO comment in the file with the issue URL
func (c *Client) UpdateCommentInFile(ctx context.Context, comment core.TodoComment, prNumber int, branch string) error {
	fmt.Printf("Updating comment in file %s (line %d) for branch %s\n", comment.FilePath, comment.LineNumber, branch)
	
	// Get file content from the PR branch
	fileContent, _, resp, err := c.client.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		comment.FilePath,
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err != nil {
		fmt.Printf("Error getting file contents: %v\n", err)
		if resp != nil {
			fmt.Printf("Response status: %s\n", resp.Status)
		}
		return fmt.Errorf("failed to get content of %s (branch: %s): %w", comment.FilePath, branch, err)
	}

	// Decode file content
	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("failed to decode content of %s: %w", comment.FilePath, err)
	}

	// Update the TODO comment with the issue URL
	lines := strings.Split(content, "\n")
	if comment.LineNumber-1 < 0 || comment.LineNumber >= len(lines) {
		return fmt.Errorf("line number %d is out of range for file %s", comment.LineNumber, comment.FilePath)
	}

	// Find the starting line of the TODO comment
	todoLineIndex := comment.LineNumber - 1
	lang := core.GetLanguageForFile(comment.FilePath)
	if lang == nil {
		return fmt.Errorf("unsupported file type: %s", comment.FilePath)
	}

	fmt.Printf("TODO line content: %s\n", lines[todoLineIndex])

	// Insert the Issue line after the TODO line
	if !strings.Contains(lines[todoLineIndex], "Issue:") {
		issueComment := fmt.Sprintf("%s Issue: %s", lang.LineComment, comment.IssueURL)
		fmt.Printf("Adding issue URL line: %s\n", issueComment)

		// Insert the issue line after the TODO line
		updatedLines := append(lines[:todoLineIndex+1], append([]string{issueComment}, lines[todoLineIndex+1:]...)...)
		updatedContent := strings.Join(updatedLines, "\n")

		// Create a commit to update the file
		sha := fileContent.GetSHA()
		message := fmt.Sprintf("Update TODO comment with issue URL in %s", comment.FilePath)
		_, resp, err = c.client.Repositories.UpdateFile(
			ctx,
			c.owner,
			c.repo,
			comment.FilePath,
			&github.RepositoryContentFileOptions{
				Message: &message,
				Content: []byte(updatedContent),
				SHA:     &sha,
				Branch:  &branch,
			},
		)
		if err != nil {
			fmt.Printf("Error updating file: %v\n", err)
			if resp != nil {
				fmt.Printf("Response status: %s\n", resp.Status)
			}
			return fmt.Errorf("failed to update file %s: %w", comment.FilePath, err)
		}
		fmt.Printf("Successfully updated file %s with issue URL\n", comment.FilePath)
	} else {
		fmt.Printf("Issue URL already exists in comment, skipping update\n")
	}

	return nil
}
