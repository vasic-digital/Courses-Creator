package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/course-creator/core-processor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiagramProcessor_NewWithConfig(t *testing.T) {
	t.Run("creates processor with custom config", func(t *testing.T) {
		config := DiagramConfig{
			Width:      800,
			Height:     600,
			Quality:    85,
			OutputDir:  "/tmp/test_diagrams",
			CacheDir:   "/tmp/test_cache",
			TempDir:    "/tmp/test_temp",
			Timeout:    30 * time.Second,
			MaxRetries: 1,
			FontPath:   "/test/font.ttf",
		}

		storage := NewMockFileStorage()
		dp := NewDiagramProcessorWithConfig(config, storage)

		assert.Equal(t, config.Width, dp.config.Width)
		assert.Equal(t, config.Height, dp.config.Height)
		assert.Equal(t, config.Quality, dp.config.Quality)
		assert.Equal(t, config.OutputDir, dp.config.OutputDir)
		assert.Equal(t, config.Timeout, dp.config.Timeout)
		assert.Equal(t, config.MaxRetries, dp.config.MaxRetries)
		assert.Equal(t, config.FontPath, dp.config.FontPath)

		assert.DirExists(t, config.OutputDir)
		assert.DirExists(t, config.CacheDir)
		assert.DirExists(t, config.TempDir)

		os.RemoveAll(config.OutputDir)
		os.RemoveAll(config.CacheDir)
		os.RemoveAll(config.TempDir)
	})

	t.Run("creates processor with default config", func(t *testing.T) {
		storage := NewMockFileStorage()
		dp := NewDiagramProcessor(storage)

		assert.Equal(t, 1920, dp.config.Width)
		assert.Equal(t, 1080, dp.config.Height)
		assert.Equal(t, 90, dp.config.Quality)
		assert.Equal(t, "/tmp/diagrams", dp.config.OutputDir)
		assert.Equal(t, 120*time.Second, dp.config.Timeout)
		assert.Equal(t, 2, dp.config.MaxRetries)
		assert.Contains(t, dp.config.FontPath, "Helvetica")

		assert.DirExists(t, dp.config.OutputDir)
		assert.DirExists(t, dp.config.CacheDir)
		assert.DirExists(t, dp.config.TempDir)

		os.RemoveAll(dp.config.OutputDir)
		os.RemoveAll(dp.config.CacheDir)
		os.RemoveAll(dp.config.TempDir)
	})
}

func TestDiagramProcessor_GenerateDiagram(t *testing.T) {
	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)
	ctx := context.Background()
	options := TestProcessingOptions()

	tests := []struct {
		name    string
		request DiagramRequest
	}{
		{
			name: "generate mermaid diagram",
			request: DiagramRequest{
				Type:    DiagramFlowchart,
				Title:   "Test Flowchart",
				Content: "graph TD\n    A --> B",
				Style:   "mermaid",
				Options: map[string]interface{}{
					"format": "png",
					"theme":  "default",
				},
			},
		},
		{
			name: "generate text-based diagram",
			request: DiagramRequest{
				Type:    DiagramProcess,
				Title:   "Test Process",
				Content: "Step 1: Start\nStep 2: Process\nStep 3: End",
				Style:   "generated",
				Options: map[string]interface{}{
					"format":      "png",
					"auto_layout": true,
				},
			},
		},
		{
			name: "generate default diagram",
			request: DiagramRequest{
				Type:    DiagramConcept,
				Title:   "Test Concept",
				Content: "Some concept description",
				Style:   "unknown",
				Options: map[string]interface{}{
					"format": "png",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram, err := dp.generateDiagram(ctx, tt.request, "test_diagram", options)
			require.NoError(t, err, "should generate diagram without error")

			assert.NotEmpty(t, diagram.ID, "should have generated ID")
			assert.Equal(t, string(tt.request.Type), diagram.Type)
			assert.Equal(t, tt.request.Title, diagram.Title)
			assert.NotNil(t, diagram.ImageURL, "should have image URL")
			assert.NotNil(t, diagram.Data, "should have data")
			assert.NotZero(t, diagram.CreatedAt, "should have creation timestamp")

			storagePath := *diagram.ImageURL
			content, exists := storage.GetFileContent(storagePath)
			assert.True(t, exists, "diagram should be saved to storage")
			assert.NotEmpty(t, content, "stored content should not be empty")
		})
	}
}

func TestDiagramProcessor_EdgeCases(t *testing.T) {
	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)
	ctx := context.Background()
	options := TestProcessingOptions()

	t.Run("handles empty content", func(t *testing.T) {
		requests := dp.detectDiagrams("")
		assert.Equal(t, 0, len(requests), "should handle empty content")

		diagrams, err := dp.ProcessDiagrams(ctx, "", options)
		require.NoError(t, err, "should handle empty content without error")
		assert.Equal(t, 0, len(diagrams), "should return empty slice for empty content")
	})

	t.Run("handles malformed mermaid code", func(t *testing.T) {
		content := "# Test\n```mermaid\ninvalid mermaid syntax\n```"

		diagrams, err := dp.ProcessDiagrams(ctx, content, options)
		require.NoError(t, err, "should handle malformed mermaid without error")
		assert.NotNil(t, diagrams)
	})

	t.Run("handles very long content", func(t *testing.T) {
		longContent := "# Test\n"
		for i := 0; i < 1000; i++ {
			longContent += "Line " + string(rune('A'+(i%26))) + "\n"
		}
		longContent += "```mermaid\ngraph TD\nA --> B\n```"

		diagrams, err := dp.ProcessDiagrams(ctx, longContent, options)
		require.NoError(t, err, "should handle long content without error")
		assert.Equal(t, 1, len(diagrams), "should detect diagram in long content")
	})

	t.Run("handles special characters in titles", func(t *testing.T) {
		content := "# Test\n## Diagram with Special Chars: Test & More < > \" ' :\nThis is a test diagram with special characters."

		requests := dp.detectDiagrams(content)
		assert.Equal(t, 1, len(requests), "should detect diagram with special chars")
		if len(requests) > 0 {
			assert.Equal(t, "Diagram with Special Chars: Test & More < > \" ' :", requests[0].Title)
		}
	})

	t.Run("handles storage failure gracefully", func(t *testing.T) {
		failingStorage := NewMockFileStorage()
		failingStorage.SetFailure("diagrams/flowchart/diagram_1.png")
		dpWithFailingStorage := NewDiagramProcessor(failingStorage)

		content := "# Test Course\n```mermaid\nflowchart TD\n    A --> B\n```"

		diagrams, err := dpWithFailingStorage.ProcessDiagrams(ctx, content, options)
		require.NoError(t, err, "should handle storage failure gracefully")
		assert.Equal(t, 0, len(diagrams), "should not return diagrams when storage fails")
	})
}

func TestDiagramProcessor_Integration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "1" {
		t.Skip("Skipping integration test")
	}

	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)
	ctx := context.Background()
	options := TestProcessingOptions()

	t.Run("end-to-end diagram processing", func(t *testing.T) {
		content := TestMarkdownContent() + "\n\n## Course Architecture:\n```mermaid\ngraph TB\n    Student -->|Access| Platform\n    Platform -->|Uses| Database\n    Platform -->|Calls| API\n    API -->|Processes| Services\n```\n\n## Learning Path:\nThe learning path consists of:\n1. Foundation concepts\n2. Core skills development  \n3. Advanced topics\n4. Project work\n5. Assessment"

		diagrams, err := dp.ProcessDiagrams(ctx, content, options)
		require.NoError(t, err, "should process diagrams end-to-end")
		assert.Equal(t, 2, len(diagrams), "should generate both diagrams")

		for _, diagram := range diagrams {
			assert.NotNil(t, diagram.ImageURL)
			content, exists := storage.GetFileContent(*diagram.ImageURL)
			assert.True(t, exists, "diagram should be saved to storage")
			assert.NotEmpty(t, content, "stored image should not be empty")
		}

		os.RemoveAll(dp.config.OutputDir)
		os.RemoveAll(dp.config.CacheDir)
		os.RemoveAll(dp.config.TempDir)
	})
}

func TestDiagramProcessor_HelperMethods(t *testing.T) {
	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)

	t.Run("getStoragePath generates correct paths", func(t *testing.T) {
		request := DiagramRequest{
			Type:  DiagramFlowchart,
			Title: "Test Diagram",
		}

		path := dp.getStoragePath(request, "test_123")
		expected := "diagrams/flowchart/test_123.png"
		assert.Equal(t, expected, path, "should generate correct storage path")
	})

	t.Run("generateDescription creates meaningful descriptions", func(t *testing.T) {
		request := DiagramRequest{
			Type:  DiagramFlowchart,
			Title: "System Architecture",
		}

		description := dp.generateDescription(request)
		expected := "A flowchart diagram showing System Architecture"
		assert.Equal(t, expected, description, "should generate correct description")
	})
}

func TestDiagramProcessor_DrawingMethods(t *testing.T) {
	testDir := "/tmp/test_diagram_drawing"
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	storage := NewMockFileStorage()
	config := DiagramConfig{
		Width:      400,
		Height:     300,
		Quality:    90,
		OutputDir:  testDir,
		CacheDir:   testDir,
		TempDir:    testDir,
		Timeout:    10 * time.Second,
		MaxRetries: 1,
		FontPath:   "/System/Library/Fonts/Helvetica.ttc",
	}

	dp := NewDiagramProcessorWithConfig(config, storage)

	t.Run("createPlaceholderDiagram generates image file", func(t *testing.T) {
		request := DiagramRequest{
			Type:    DiagramFlowchart,
			Title:   "Test Flowchart",
			Content: "Test content",
		}

		outputPath := filepath.Join(testDir, "test_placeholder.png")
		imagePath, err := dp.createPlaceholderDiagram(outputPath, request, "Test")
		require.NoError(t, err, "should create placeholder diagram")
		assert.Equal(t, outputPath, imagePath, "should return correct path")

		_, err = os.Stat(imagePath)
		assert.NoError(t, err, "image file should exist")

		fileInfo, _ := os.Stat(imagePath)
		assert.True(t, fileInfo.Size() > 0, "image file should not be empty")
	})

	t.Run("createTypedDiagram generates image file", func(t *testing.T) {
		request := DiagramRequest{
			Type:    DiagramFlowchart,
			Title:   "Test Flowchart",
			Content: "Step 1\nStep 2\nStep 3",
		}

		outputPath := filepath.Join(testDir, "test_typed.png")
		imagePath, err := dp.createTypedDiagram(outputPath, request)
		require.NoError(t, err, "should create typed diagram")
		assert.Equal(t, outputPath, imagePath, "should return correct path")

		_, err = os.Stat(imagePath)
		assert.NoError(t, err, "image file should exist")

		fileInfo, _ := os.Stat(imagePath)
		assert.True(t, fileInfo.Size() > 0, "image file should not be empty")
	})
}

func TestDiagramProcessor_TextBasedDiagramDetection(t *testing.T) {
	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "detects text-based diagram with colon",
			content:  "# Test\n## Process Diagram:\nStep 1: Start\nStep 2: Process\nStep 3: End",
			expected: 1,
		},
		{
			name:     "detects text-based diagram without colon",
			content:  "# Test\n## Architecture Diagram\nSystem components:\n- Frontend\n- Backend\n- Database",
			expected: 1,
		},
		{
			name:     "does not detect regular sections",
			content:  "# Test\n## Regular Section\nThis is just regular text without diagram indicators.",
			expected: 0,
		},
		{
			name:     "detects diagram in mixed content",
			content:  "# Test\n## Introduction\nSome intro text.\n\n## Flow Diagram:\nA -> B -> C\n\n## Conclusion\nFinal thoughts.",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := dp.detectDiagrams(tt.content)
			assert.Equal(t, tt.expected, len(requests), "should detect correct number of text-based diagrams")
		})
	}
}

func TestDiagramProcessor_AnalyzeMermaidDiagram_Detailed(t *testing.T) {
	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)

	tests := []struct {
		name     string
		content  string
		expected DiagramType
		title    string
	}{
		{
			name:     "flowchart with graph keyword",
			content:  "title: My Flowchart\ngraph TD\n    A --> B",
			expected: DiagramFlowchart,
			title:    "My Flowchart",
		},
		{
			name:     "flowchart with flowchart keyword",
			content:  "title: Process Flow\nflowchart LR\n    Start --> End",
			expected: DiagramFlowchart,
			title:    "Process Flow",
		},
		{
			name:     "sequence diagram",
			content:  "title: API Sequence\nsequenceDiagram\n    Client->>Server: Request",
			expected: DiagramSequence,
			title:    "API Sequence",
		},
		{
			name:     "class diagram",
			content:  "title: Class Model\nclassDiagram\n    Animal <|-- Dog",
			expected: DiagramClass,
			title:    "Class Model",
		},
		{
			name:     "entity relationship diagram",
			content:  "title: Database ERD\nerDiagram\n    USER ||--o{ POST : writes",
			expected: DiagramEntity,
			title:    "Database ERD",
		},
		{
			name:     "mind map",
			content:  "title: Project Plan\nmindmap\n  root((Ideas))",
			expected: DiagramMindMap,
			title:    "Project Plan",
		},
		{
			name:     "architecture diagram with graph TB",
			content:  "graph TB\n    A --> B",
			expected: DiagramArchitecture,
			title:    "Untitled Diagram",
		},
		{
			name:     "architecture diagram with graph TD",
			content:  "graph TD\n    A --> B",
			expected: DiagramArchitecture,
			title:    "Untitled Diagram",
		},
		{
			name:     "unknown diagram type",
			content:  "customDiagram\n    node1 -- node2",
			expected: DiagramConcept,
			title:    "Untitled Diagram",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagramType, title := dp.analyzeMermaidDiagram(tt.content)
			assert.Equal(t, tt.expected, diagramType, "should identify correct diagram type")
			assert.Equal(t, tt.title, title, "should extract correct title")
		})
	}
}
