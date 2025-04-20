# pdd-action
Github Action to add Puzzle Driven Development into your Github Repository 

// TODO: Add comprehensive documentation
// Labels: documentation,enhancement
// This project needs more comprehensive documentation including:
// - Detailed examples for different languages
// - Troubleshooting section
// - Advanced configuration options

Puzzle Driven Development (PDD) is a software development methodology that focuses on breaking down complex problems into smaller, manageable puzzles.
It encourages collaboration, creativity, and iterative problem-solving to deliver high-quality software solutions. 
PDD emphasizes the importance of understanding the problem domain and leveraging the collective intelligence of the development team to find innovative solutions.

How process works:
1. Developer start working on github issue
2. In many cases, issue doesn't cover all underlying complexity
3. Along Development process, developer creates TODO comments in the codebase, to highlight the additional work that needs to be done 
2. This tool will parse PR and find all TODO comments
3. Then tool will create new issues in the repository, based on the TODO comments
4. and update comments in the PR with ids of the new issues

Comments format before issue creation:
```
// TODO: {issue_title}
// Labels: {comma_separated_labels} (optional)
// {issue_description}
// {issue_description_continue}
```

Comments format after issue creation:
```
// TODO: {issue_title}
// Issue: {issue_url}
// Labels: {comma_separated_labels} (optional)
// {issue_description}
// {issue_description_continue}
```

This tool supports comments format for as many languages as possible, including:
GoLang, Java, Python, JavaScript, TypeScript, C#, C++, C, Ruby, Swift, Kotlin, Rust, PHP, HTML, CSS, Shell Script, Bash Script, PowerShell Script, SQL, R, Perl, Haskell, Scala, Groovy, Lua, Elixir, Erlang, F#, Objective-C

## Usage

To use this action in your workflow, add the following YAML to your GitHub workflow configuration file (e.g., `.github/workflows/pdd.yml`):

```yaml
name: PDD Action

on:
  pull_request:
    types: [closed]

permissions:
  contents: write
  issues: write
  pull-requests: write

jobs:
  pdd:
    if: github.event.pull_request.merged == true
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
        
      - name: Run PDD Action
        uses: ksysoev/pdd-action@v1
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          branch_name: main
          issue_title_prefix: "[PDD]"
```

## Configuration

The following inputs are available:

| Input | Description | Required | Default |
| ----- | ----------- | -------- | ------- |
| `github_token` | GitHub token to create issues in the repository | Yes | N/A |
| `branch_name` | Branch name to create issues in the repository | No | `main` |
| `issue_title_prefix` | Prefix to add to issue titles | No | `` |

## How It Works

1. The action runs when a pull request is merged to the specified branch.
2. It scans the codebase for TODO comments in the specified format.
3. For each unprocessed TODO comment (comments without an associated issue URL), it creates a new GitHub issue.
4. It then updates the TODO comment in the code with the issue URL.

> **Important:** Make sure to set the appropriate permissions in your workflow file as shown in the example above. The action needs `contents: write`, `issues: write`, and `pull-requests: write` permissions to function correctly.

// TODO: Add section on supported comment formats
// Labels: documentation
// Provide examples of TODO comments in different languages
// to make it clearer how to use the tool across different codebases

## Development

To build and test this action locally:

```bash
# Clone the repository
git clone https://github.com/ksysoev/pdd-action.git
cd pdd-action

# Build the project
go build -o pdd-action ./cmd/pdd-action

# Run the project locally
./pdd-action
```