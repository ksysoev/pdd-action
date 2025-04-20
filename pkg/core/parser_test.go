package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLanguageForFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "Go file",
			filename: "sample.go",
			want:     true,
		},
		{
			name:     "JavaScript file",
			filename: "sample.js",
			want:     true,
		},
		{
			name:     "Python file",
			filename: "sample.py",
			want:     true,
		},
		{
			name:     "Unsupported file",
			filename: "sample.xyz",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := GetLanguageForFile(tt.filename)
			if tt.want {
				assert.NotNil(t, lang, "Expected language to be found for %s", tt.filename)
			} else {
				assert.Nil(t, lang, "Expected no language to be found for %s", tt.filename)
			}
		})
	}
}

func TestParseTodoComments(t *testing.T) {
	// Create a temporary file with TODO comments
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "sample.go")

	content := `package sample

// Regular comment

// TODO: Sample task
// Labels: enhancement,bug
// This is a description
// Spanning multiple lines

func sampleFunc() {
	// Another comment
	
	// TODO: Another task
	// This is another description
}`

	err := os.WriteFile(tempFile, []byte(content), 0644)
	assert.NoError(t, err)

	// Parse the file
	comments, err := ParseTodoComments(tempFile)
	assert.NoError(t, err)
	assert.Len(t, comments, 2, "Expected to find 2 TODO comments")

	// Check the first comment
	assert.Equal(t, "Sample task", comments[0].Title)
	assert.Equal(t, []string{"enhancement", "bug"}, comments[0].Labels)
	assert.Equal(t, []string{"This is a description", "Spanning multiple lines"}, comments[0].Description)

	// Check the second comment
	assert.Equal(t, "Another task", comments[1].Title)
	assert.Empty(t, comments[1].Labels)
	assert.Equal(t, []string{"This is another description"}, comments[1].Description)
}
