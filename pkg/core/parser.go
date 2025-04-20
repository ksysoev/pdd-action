package core

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Language defines comment styles for different programming languages
type Language struct {
	Extensions        []string
	LineComment       string
	BlockCommentStart string
	BlockCommentEnd   string
}

var supportedLanguages = []Language{
	{Extensions: []string{".go"}, LineComment: "//"},
	{Extensions: []string{".java", ".js", ".ts", ".jsx", ".tsx", ".c", ".cpp", ".cs", ".h", ".hpp", ".swift", ".kt", ".rs", ".php", ".scala", ".groovy"}, LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	{Extensions: []string{".py", ".rb", ".pl", ".r", ".sh", ".bash"}, LineComment: "#"},
	{Extensions: []string{".lua"}, LineComment: "--"},
	{Extensions: []string{".sql"}, LineComment: "--", BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	{Extensions: []string{".html", ".xml"}, BlockCommentStart: "<!--", BlockCommentEnd: "-->"},
	{Extensions: []string{".css"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	{Extensions: []string{".ex", ".exs"}, LineComment: "#"},
	{Extensions: []string{".erl", ".hrl"}, LineComment: "%"},
	{Extensions: []string{".hs"}, LineComment: "--", BlockCommentStart: "{-", BlockCommentEnd: "-}"},
	{Extensions: []string{".ps1"}, LineComment: "#"},
	{Extensions: []string{".fs"}, LineComment: "//", BlockCommentStart: "(*", BlockCommentEnd: "*)"},
	{Extensions: []string{".m"}, LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	{Extensions: []string{".md", ".markdown"}, LineComment: "//"},
}

// GetLanguageForFile determines the language of a file based on its extension
func GetLanguageForFile(filename string) *Language {
	ext := filepath.Ext(filename)
	for _, lang := range supportedLanguages {
		for _, langExt := range lang.Extensions {
			if ext == langExt {
				return &lang
			}
		}
	}
	return nil
}

// ParseTodoComments scans a file for TODO comments in the specified format
func ParseTodoComments(filePath string) ([]TodoComment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lang := GetLanguageForFile(filePath)
	if lang == nil {
		return nil, nil // Unsupported file type
	}

	var comments []TodoComment
	scanner := bufio.NewScanner(file)

	todoRegex := regexp.MustCompile(`TODO:(.+)`)
	labelsRegex := regexp.MustCompile(`Labels:(.+)`)
	issueRegex := regexp.MustCompile(`Issue:(.+)`)

	lineNum := 0
	var currentComment *TodoComment
	var inBlockComment bool

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Handle block comments start/end
		if lang.BlockCommentStart != "" && strings.Contains(trimmedLine, lang.BlockCommentStart) {
			inBlockComment = true
		}
		if lang.BlockCommentEnd != "" && strings.Contains(trimmedLine, lang.BlockCommentEnd) {
			inBlockComment = false
			continue
		}

		// Check if the line is a comment
		isComment := false
		commentContent := ""

		if lang.LineComment != "" && strings.HasPrefix(trimmedLine, lang.LineComment) {
			isComment = true
			commentContent = strings.TrimSpace(strings.TrimPrefix(trimmedLine, lang.LineComment))
		} else if inBlockComment {
			isComment = true
			commentContent = trimmedLine
		}

		if !isComment {
			// If we were collecting a comment and found a non-comment line, finalize the current comment
			if currentComment != nil {
				comments = append(comments, *currentComment)
				currentComment = nil
			}
			continue
		}

		// Process comment content
		if todoMatch := todoRegex.FindStringSubmatch(commentContent); todoMatch != nil && currentComment == nil {
			// Start a new TODO comment
			title := strings.TrimSpace(todoMatch[1])
			currentComment = &TodoComment{
				FilePath:   filePath,
				LineNumber: lineNum,
				Title:      title,
			}
		} else if currentComment != nil {
			// Check for existing issue URL
			if issueMatch := issueRegex.FindStringSubmatch(commentContent); issueMatch != nil {
				currentComment.IssueURL = strings.TrimSpace(issueMatch[1])
			} else if labelsMatch := labelsRegex.FindStringSubmatch(commentContent); labelsMatch != nil {
				// Extract labels
				labelsStr := strings.TrimSpace(labelsMatch[1])
				labels := strings.Split(labelsStr, ",")
				for i, label := range labels {
					labels[i] = strings.TrimSpace(label)
				}
				currentComment.Labels = labels
			} else if commentContent != "" {
				// Add to description if not a special directive
				currentComment.Description = append(currentComment.Description, commentContent)
			}
		}
	}

	// Add the last comment if there is one
	if currentComment != nil {
		comments = append(comments, *currentComment)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// ScanDirectory recursively scans a directory for TODO comments
func ScanDirectory(dir string, excludeDirs []string) ([]TodoComment, error) {
	var allComments []TodoComment

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			for _, excludeDir := range excludeDirs {
				if strings.HasPrefix(path, excludeDir) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Parse files
		comments, err := ParseTodoComments(path)
		if err != nil {
			return err
		}

		allComments = append(allComments, comments...)
		return nil
	})

	return allComments, err
}

// FilterUnprocessedComments returns comments that don't have an issue URL
func FilterUnprocessedComments(comments []TodoComment) []TodoComment {
	var unprocessed []TodoComment
	for _, comment := range comments {
		if comment.IssueURL == "" {
			unprocessed = append(unprocessed, comment)
		}
	}
	return unprocessed
}
