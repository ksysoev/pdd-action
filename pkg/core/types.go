package core

// TodoComment represents a parsed TODO comment from code
type TodoComment struct {
	FilePath    string
	LineNumber  int
	Title       string
	Description []string
	Labels      []string
	IssueURL    string
}

// Config represents the GitHub Action configuration
type Config struct {
	GitHubToken      string
	BranchName       string
	IssueTitlePrefix string
}
