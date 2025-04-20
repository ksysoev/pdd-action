package github

import (
	"context"
	"fmt"
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
	owner := parts[0]
	repo := parts[1]

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

		// Create the issue
		issue, _, err := c.client.Issues.Create(ctx, c.owner, c.repo, &github.IssueRequest{
			Title:  &title,
			Body:   &body,
			Labels: &comment.Labels,
		})
		if err != nil {
			return processedComments, fmt.Errorf("failed to create issue for comment in %s (line %d): %w",
				comment.FilePath, comment.LineNumber, err)
		}

		// Update the comment with the issue URL
		comment.IssueURL = issue.GetHTMLURL()
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
	// Get file content from the PR branch
	fileContent, _, _, err := c.client.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		comment.FilePath,
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err != nil {
		return fmt.Errorf("failed to get content of %s: %w", comment.FilePath, err)
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

	// Insert the Issue line after the TODO line
	if !strings.Contains(lines[todoLineIndex], "Issue:") {
		issueComment := fmt.Sprintf("%s Issue: %s", lang.LineComment, comment.IssueURL)

		// Insert the issue line after the TODO line
		updatedLines := append(lines[:todoLineIndex+1], append([]string{issueComment}, lines[todoLineIndex+1:]...)...)
		updatedContent := strings.Join(updatedLines, "\n")

		// Create a commit to update the file
		sha := fileContent.GetSHA()
		message := fmt.Sprintf("Update TODO comment with issue URL in %s", comment.FilePath)
		_, _, err = c.client.Repositories.UpdateFile(
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
			return fmt.Errorf("failed to update file %s: %w", comment.FilePath, err)
		}
	}

	return nil
}
