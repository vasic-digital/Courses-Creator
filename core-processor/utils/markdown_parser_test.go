package utils_test

import (
	"testing"

	"github.com/course-creator/core-processor/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMarkdownParser(t *testing.T) {
	parser := utils.NewMarkdownParser()
	assert.NotNil(t, parser)
}

func TestMarkdownParser_Parse_Simple(t *testing.T) {
	parser := utils.NewMarkdownParser()

	markdown := "# Introduction to Go\n\nThis is a course about Go programming language.\n\n## Getting Started\n\nInstall Go from the official website.\n\n## Basic Syntax\n\nLearn about variables and functions."

	result, err := parser.Parse(markdown)
	require.NoError(t, err)
	assert.Equal(t, "Introduction to Go", result.Title)
	assert.Equal(t, "This is a course about Go programming language.", result.Description)
	assert.Len(t, result.Sections, 2)

	assert.Equal(t, "Getting Started", result.Sections[0].Title)
	assert.Equal(t, "Install Go from the official website.", result.Sections[0].Content)
	assert.Equal(t, 0, result.Sections[0].Order)

	assert.Equal(t, "Basic Syntax", result.Sections[1].Title)
	assert.Equal(t, "Learn about variables and functions.", result.Sections[1].Content)
	assert.Equal(t, 1, result.Sections[1].Order)

	assert.NotNil(t, result.Metadata)
}

func TestMarkdownParser_Parse_NoTitle(t *testing.T) {
	parser := utils.NewMarkdownParser()

	markdown := "This is content without a title header.\n\n## First Section\n\nContent here."

	result, err := parser.Parse(markdown)
	require.NoError(t, err)
	// First header becomes the title
	assert.Equal(t, "First Section", result.Title)
	// Content before and after first header (until next header) becomes description
	assert.Equal(t, "This is content without a title header.\nContent here.", result.Description)
	// No sections since first header was used as title
	assert.Len(t, result.Sections, 0)
}

func TestMarkdownParser_Parse_WithImages(t *testing.T) {
	parser := utils.NewMarkdownParser()

	markdown := "# Course with Images\n\n## Section with Image\n\nHere is an image: ![Alt text](https://example.com/image1.png)\n\nAnother image: ![Another alt](https://example.com/image2.jpg)"

	result, err := parser.Parse(markdown)
	require.NoError(t, err)
	assert.Equal(t, "Course with Images", result.Title)
	assert.Len(t, result.Sections, 1)

	section := result.Sections[0]
	assert.Equal(t, "Section with Image", section.Title)
	assert.Contains(t, section.Content, "Here is an image:")
	assert.Len(t, section.Images, 2)
	assert.Equal(t, "https://example.com/image1.png", section.Images[0])
	assert.Equal(t, "https://example.com/image2.jpg", section.Images[1])
}

func TestMarkdownParser_Parse_WithCodeBlocks(t *testing.T) {
	parser := utils.NewMarkdownParser()

	// Use raw string literal for code blocks
	markdown := `# Go Programming

## Hello World

Here's a simple Go program:

` + "```" + `go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

That's it!`

	result, err := parser.Parse(markdown)
	require.NoError(t, err)
	assert.Equal(t, "Go Programming", result.Title)
	assert.Len(t, result.Sections, 1)

	section := result.Sections[0]
	assert.Equal(t, "Hello World", section.Title)
	assert.Contains(t, section.Content, "Here's a simple Go program:")
	assert.Contains(t, section.Content, "package main")
}

func TestMarkdownParser_Parse_EmptyContent(t *testing.T) {
	parser := utils.NewMarkdownParser()

	result, err := parser.Parse("")
	require.NoError(t, err)
	assert.Equal(t, "Untitled Course", result.Title)
	assert.Equal(t, "", result.Description)
	assert.Len(t, result.Sections, 0)
}

func TestMarkdownParser_Parse_OnlyTitle(t *testing.T) {
	parser := utils.NewMarkdownParser()

	markdown := "# Only Title Here"

	result, err := parser.Parse(markdown)
	require.NoError(t, err)
	assert.Equal(t, "Only Title Here", result.Title)
	assert.Equal(t, "", result.Description)
	assert.Len(t, result.Sections, 0)
}

func TestMarkdownParser_Parse_MultipleLevelHeaders(t *testing.T) {
	parser := utils.NewMarkdownParser()

	markdown := "# Main Title\n\n## Section 1\n\nContent 1\n\n### Subsection 1.1\n\nContent 1.1\n\n## Section 2\n\nContent 2"

	result, err := parser.Parse(markdown)
	require.NoError(t, err)
	assert.Equal(t, "Main Title", result.Title)
	// All headers (##, ###) become separate sections
	assert.Len(t, result.Sections, 3)

	assert.Equal(t, "Section 1", result.Sections[0].Title)
	assert.Equal(t, "Content 1", result.Sections[0].Content)
	assert.Equal(t, 0, result.Sections[0].Order)

	assert.Equal(t, "Subsection 1.1", result.Sections[1].Title)
	assert.Equal(t, "Content 1.1", result.Sections[1].Content)
	assert.Equal(t, 1, result.Sections[1].Order)

	assert.Equal(t, "Section 2", result.Sections[2].Title)
	assert.Equal(t, "Content 2", result.Sections[2].Content)
	assert.Equal(t, 2, result.Sections[2].Order)
}

func TestMarkdownParser_ExtractTitle(t *testing.T) {
	parser := utils.NewMarkdownParser()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple title",
			content:  "# My Course Title\n\nContent here",
			expected: "My Course Title",
		},
		{
			name:     "title with extra spaces",
			content:  "#   My Course Title   \n\nContent",
			expected: "My Course Title",
		},
		{
			name:     "no title",
			content:  "Just content\nNo title here",
			expected: "Untitled Course",
		},
		{
			name:     "title not at beginning",
			content:  "Some intro\n# Actual Title\nMore content",
			expected: "Actual Title",
		},
		{
			name:     "multiple hashes",
			content:  "### Small Title\nContent",
			expected: "Small Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Title)
		})
	}
}

func TestMarkdownParser_ExtractDescription(t *testing.T) {
	parser := utils.NewMarkdownParser()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple description",
			content:  "# Title\n\nThis is the description.\n\n## Section 1",
			expected: "This is the description.",
		},
		{
			name:     "multi-line description",
			content:  "# Title\n\nLine 1 of description.\nLine 2 of description.\n\n## Section",
			expected: "Line 1 of description.\nLine 2 of description.",
		},
		{
			name:     "no description",
			content:  "# Title\n\n## Section",
			expected: "",
		},
		{
			name:     "description with blank lines",
			content:  "# Title\n\n\nDescription here.\n\n\n## Section",
			expected: "Description here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Description)
		})
	}
}
