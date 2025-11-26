package contract

import (
	"testing"

	"github.com/course-creator/core-processor/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownParser_Parse_ValidMarkdown(t *testing.T) {
	parser := utils.NewMarkdownParser()

	content := `# Course Title

This is the course description.

## Section 1

This is the first section content.

## Section 2

This is the second section content.
`

	parsed, err := parser.Parse(content)
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "Course Title", parsed.Title)
	assert.Contains(t, parsed.Description, "course description")
	assert.Len(t, parsed.Sections, 2)
	assert.Equal(t, "Section 1", parsed.Sections[0].Title)
	assert.Equal(t, "Section 2", parsed.Sections[1].Title)
}

func TestMarkdownParser_Parse_EmptyContent(t *testing.T) {
	parser := utils.NewMarkdownParser()

	content := ""

	parsed, err := parser.Parse(content)
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "Untitled Course", parsed.Title)
	assert.Len(t, parsed.Sections, 0)
}

func TestMarkdownParser_Parse_NoHeaders(t *testing.T) {
	parser := utils.NewMarkdownParser()

	content := `This is just some content without headers.`

	parsed, err := parser.Parse(content)
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, "Untitled Course", parsed.Title)
	assert.Equal(t, content, parsed.Description)
	assert.Len(t, parsed.Sections, 0)
}
