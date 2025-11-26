package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCourseGenerator_GenerateCourse_BasicMarkdown(t *testing.T) {
	// Create temporary markdown file
	tempDir := t.TempDir()
	markdownPath := filepath.Join(tempDir, "test.md")
	outputDir := filepath.Join(tempDir, "output")

	content := `# Test Course

This is a test course.

## Introduction

Welcome to the course!

## Main Content

This is the main content section.
`

	err := os.WriteFile(markdownPath, []byte(content), 0644)
	require.NoError(t, err)

	// Create course generator
	generator := pipeline.NewCourseGenerator()

	// Generate course
	options := models.ProcessingOptions{
		Quality: "standard",
	}

	course, err := generator.GenerateCourse(markdownPath, outputDir, options)
	require.NoError(t, err)
	assert.NotNil(t, course)
	assert.Equal(t, "Test Course", course.Title)
	assert.Len(t, course.Lessons, 2)
	assert.Equal(t, "Introduction", course.Lessons[0].Title)
	assert.Equal(t, "Main Content", course.Lessons[1].Title)
}

func TestCourseGenerator_GenerateCourse_EmptyMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	markdownPath := filepath.Join(tempDir, "empty.md")
	outputDir := filepath.Join(tempDir, "output")

	content := ""

	err := os.WriteFile(markdownPath, []byte(content), 0644)
	require.NoError(t, err)

	generator := pipeline.NewCourseGenerator()
	options := models.ProcessingOptions{}

	course, err := generator.GenerateCourse(markdownPath, outputDir, options)
	require.NoError(t, err)
	assert.NotNil(t, course)
	assert.Equal(t, "Untitled Course", course.Title)
	assert.Len(t, course.Lessons, 0)
}
