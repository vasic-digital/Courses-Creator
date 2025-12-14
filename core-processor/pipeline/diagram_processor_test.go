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

func TestDiagramProcessor_DetectDiagrams(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "detects mermaid diagram",
			content: `# Test Course

## Introduction
This is a test course.

## Architecture Diagram
```mermaid
graph TD
    A[Client] --> B[API Gateway]
    B --> C[Service 1]
    B --> D[Service 2]
    C --> E[Database]
    D --> E
````,
			expected: 1,
		},
		{
			name: "detects multiple mermaid diagrams",
			content: `# Test Course

## Flowchart
```mermaid
flowchart TD
    Start --> Process
    Process --> Decision
    Decision -->|Yes| End
    Decision -->|No| Process
```

## Sequence Diagram
```mermaid
sequenceDiagram
    Alice->>John: Hello John, how are you?
    John-->>Alice: Great!
    Alice-)John: See you later!
````,
			expected: 2,
		},
		{
			name: "detects text-based diagram",
			content: `# Test Course

## Process Diagram:
This diagram shows the data flow process:
1. Data collection from sensors
2. Data preprocessing and cleaning
3. Feature extraction
4. Model training
5. Prediction output`,
			expected: 1,
		},
		{
			name: "detects mixed diagrams",
			content: `# Test Course

## System Architecture
```mermaid
graph TB
    Web[Web Client] --> API[API Server]
    API --> DB[(Database)]
    API --> Cache[Redis Cache]
```

## Deployment Process:
The deployment process involves these steps:
- Code commit and review
- Automated testing
- Build and package
- Deployment to staging
- Smoke tests
- Production deployment`,
			expected: 2,
		},
		{
			name: "no diagrams in content",
			content: `# Test Course

This is a simple course without any diagrams.
Just plain text content for testing purposes.`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockFileStorage()
			dp := NewDiagramProcessor(storage)
			
			requests := dp.detectDiagrams(tt.content)
			assert.Equal(t, tt.expected, len(requests), "should detect correct number of diagrams")
		})
	}
}

func TestDiagramProcessor_AnalyzeMermaidDiagram(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected DiagramType
		title    string
	}{
		{
			name: "flowchart diagram",
			content: `title: System Flow
flowchart TD
    Start --> Process
    Process --> End`,
			expected: DiagramFlowchart,
			title:    "System Flow",
		},
		{
			name: "sequence diagram",
			content: `title: User Interaction
sequenceDiagram
    User->>API: Request
    API-->>User: Response`,
			expected: DiagramSequence,
			title:    "User Interaction",
		},
		{
			name: "class diagram",
			content: `title: Class Structure
classDiagram
    Animal <|-- Duck
    Animal <|-- Fish
    Animal <|-- Zebra`,
			expected: DiagramClass,
			title:    "Class Structure",
		},
		{
			name: "entity relationship diagram",
			content: `title: Database Schema
erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains`,
			expected: DiagramEntity,
			title:    "Database Schema",
		},
		{
			name: "mind map",
			content: `title: Project Ideas
mindmap
  root((Project Ideas))
    Web Development
      E-commerce
      Blog Platform
    Mobile Apps
      Fitness Tracker
      Recipe Manager`,
			expected: DiagramMindMap,
			title:    "Project Ideas",
		},
		{
			name: "architecture diagram",
			content: `graph TB
    Client --> LoadBalancer
    LoadBalancer --> Server1
    LoadBalancer --> Server2`,
			expected: DiagramArchitecture,
			title:    "Untitled Diagram",
		},
		{
			name: "concept diagram (default)",
			content: `some unknown diagram type
with custom syntax`,
			expected: DiagramConcept,
			title:    "Untitled Diagram",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockFileStorage()
			dp := NewDiagramProcessor(storage)
			
			diagramType, title := dp.analyzeMermaidDiagram(tt.content)
			assert.Equal(t, tt.expected, diagramType, "should identify correct diagram type")
			assert.Equal(t, tt.title, title, "should extract correct title")
		})
	}
}

func TestDiagramProcessor_InferDiagramType(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected DiagramType
	}{
		{
			name:     "flowchart inference",
			content:  "This shows the flow of data through the system steps",
			expected: DiagramFlowchart,
		},
		{
			name:     "sequence inference",
			content:  "Sequence of messages between client and server interactions",
			expected: DiagramSequence,
		},
		{
			name:     "class inference",
			content:  "Class hierarchy with inheritance and methods",
			expected: DiagramClass,
		},
		{
			name:     "entity inference",
			content:  "Entity relationship database schema",
			expected: DiagramEntity,
		},
		{
			name:     "mind map inference",
			content:  "Mind map of ideas and branches",
			expected: DiagramMindMap,
		},
		{
			name:     "architecture inference",
			content:  "System architecture with components",
			expected: DiagramArchitecture,
		},
		{
			name:     "concept inference",
			content:  "Concept model of the idea",
			expected: DiagramConcept,
		},
		{
			name:     "process inference (default)",
			content:  "Some random content without keywords",
			expected: DiagramProcess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockFileStorage()
			dp := NewDiagramProcessor(storage)
			
			diagramType := dp.inferDiagramType(tt.content)
			assert.Equal(t, tt.expected, diagramType, "should infer correct diagram type")
		})
	}
}

func TestDiagramProcessor_CountElements(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "counts bullet points",
			content: `- Step 1: Collect data
- Step 2: Process data
- Step 3: Analyze results
- Step 4: Generate report`,
			expected: 4,
		},
		{
			name: "counts numbered items",
			content: `1. First phase
2. Second phase
3. Third phase`,
			expected: 3,
		},
		{
			name: "counts keywords",
			content: `The process involves multiple steps and stages with different elements and nodes in each box`,
			expected: 6, // steps, stages, elements, nodes, box (process counted twice)
		},
		{
			name: "mixed content",
			content: `Process steps:
- Step 1: Data collection
- Step 2: Processing
- Step 3: Analysis

Each stage has multiple elements and nodes.`,
			expected: 8, // 3 bullets + process + steps + stage + elements + nodes
		},
		{
			name:     "minimum elements",
			content:  "Simple content",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockFileStorage()
			dp := NewDiagramProcessor(storage)
			
			count := dp.countElements(tt.content)
			assert.Equal(t, tt.expected, count, "should count elements correctly")
		})
	}
}

func TestDiagramProcessor_ProcessDiagrams(t *testing.T) {
	storage := NewMockFileStorage()
	dp := NewDiagramProcessor(storage)
	ctx := context.Background()
	options := TestProcessingOptions()

	t.Run("processes content with diagrams", func(t *testing.T) {
		content := `# Test Course

## System Flow
```mermaid
flowchart TD
    Start --> Process
    Process --> End
```

## Deployment Process:
The deployment involves these steps:
- Code commit
- Testing
- Deployment`

		diagrams, err := dp.ProcessDiagrams(ctx, content, options)
		require.NoError(t, err, "should process diagrams without error")
		assert.Equal(t, 2, len(diagrams), "should generate two diagrams")
		
		// Verify diagram properties
		assert.Equal(t, "flowchart", diagrams[0].Type)
		assert.Equal(t, "Untitled Diagram", diagrams[0].Title)
		assert.NotNil(t, diagrams[0].ImageURL)
		
		assert.Equal(t, "process", diagrams[1].Type)
		assert.Equal(t, "Deployment Process", diagrams[1].Title)
		assert.NotNil(t, diagrams[1].ImageURL)
	})

	t.Run("handles content without diagrams", func(t *testing.T) {
		content := `# Test Course
This is simple content without any diagrams.`

		diagrams, err := dp.ProcessDiagrams(ctx, content, options)
		require.NoError(t, err, "should handle content without diagrams")
		assert.Equal(t, 0, len(diagrams), "should return empty slice")
	})

	t.Run("handles storage failure gracefully", func(t *testing.T) {
		// Create a mock storage that will fail
		failingStorage := NewMockFileStorage()
		failingStorage.SetFailure("diagrams/flowchart/diagram_1.png")
		dpWithFailingStorage := NewDiagramProcessor(failingStorage)
		
		content := `# Test Course
```mermaid
flowchart TD
    A --> B
````

		diagrams, err := dpWithFailingStorage.ProcessDiagrams(ctx, content, options)
		// The processor should handle storage failure and continue
		// (it logs error but doesn't return it from ProcessDiagrams)
		require.NoError(t, err, "should handle storage failure gracefully")
		assert.Equal(t, 0, len(diagrams), "should not return diagrams when storage fails")
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
			
			// Verify storage was called
			storagePath := *diagram.ImageURL
			content, exists := storage.GetFileContent(storagePath)
			assert.True(t, exists, "diagram should be saved to storage")
			assert.NotEmpty(t, content, "stored content should not be empty")
		})
	}
}

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
		
		// Verify directories were created
		assert.DirExists(t, config.OutputDir)
		assert.DirExists(t, config.CacheDir)
		assert.DirExists(t, config.TempDir)
		
		// Clean up test directories
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
		
		// Verify directories were created
		assert.DirExists(t, dp.config.OutputDir)
		assert.DirExists(t, dp.config.CacheDir)
		assert.DirExists(t, dp.config.TempDir)
		
		// Clean up test directories
		os.RemoveAll(dp.config.OutputDir)
		os.RemoveAll(dp.config.CacheDir)
		os.RemoveAll(dp.config.TempDir)
	})
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
		content := `# Test
```mermaid
invalid mermaid syntax
````
		
		diagrams, err := dp.ProcessDiagrams(ctx, content, options)
		// The processor should handle syntax errors gracefully
		require.NoError(t, err, "should handle malformed mermaid without error")
		// It may or may not generate a diagram - depends on implementation
		// Just ensure no panic
		assert.NotNil(t, diagrams)
	})

	t.Run("handles very long content", func(t *testing.T) {
		// Create very long content
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
		content := `# Test
## Diagram with Special Chars: Test & More < > " ' :
This is a test diagram with special characters.`
		
		requests := dp.detectDiagrams(content)
		assert.Equal(t, 1, len(requests), "should detect diagram with special chars")
		if len(requests) > 0 {
			assert.Equal(t, "Diagram with Special Chars: Test & More < > \" ' :", requests[0].Title)
		}
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
		content := TestMarkdownContent() + `

## Course Architecture:
```mermaid
graph TB
    Student -->|Access| Platform
    Platform -->|Uses| Database
    Platform -->|Calls| API
    API -->|Processes| Services
```

## Learning Path:
The learning path consists of:
1. Foundation concepts
2. Core skills development  
3. Advanced topics
4. Project work
5. Assessment`

		diagrams, err := dp.ProcessDiagrams(ctx, content, options)
		require.NoError(t, err, "should process diagrams end-to-end")
		assert.Equal(t, 2, len(diagrams), "should generate both diagrams")
		
		// Verify storage
		for _, diagram := range diagrams {
			assert.NotNil(t, diagram.ImageURL)
			content, exists := storage.GetFileContent(*diagram.ImageURL)
			assert.True(t, exists, "diagram should be saved to storage")
			assert.NotEmpty(t, content, "stored image should not be empty")
		}
		
		// Clean up
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

	t.Run("getColorForIndex returns valid colors", func(t *testing.T) {
		colors := []color.RGBA{}
		for i := 0; i < 10; i++ {
			color := dp.getColorForIndex(i)
			colors = append(colors, color)
			// Verify color has valid RGBA values
			assert.True(t, color.R >= 0 && color.R <= 255)
			assert.True(t, color.G >= 0 && color.G <= 255)
			assert.True(t, color.B >= 0 && color.B <= 255)
			assert.Equal(t, uint8(255), color.A, "alpha should be fully opaque")
		}
		
		// Verify colors cycle (index 0 and 6 should be same since len(colors)=6)
		assert.Equal(t, colors[0], colors[6], "colors should cycle")
	})
}

func TestDiagramProcessor_DrawingMethods(t *testing.T) {
	// Create a test directory for output
	testDir := "/tmp/test_diagram_drawing"
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	storage := NewMockFileStorage()
	config := DiagramConfig{
		Width:     400,
		Height:    300,
		Quality:   90,
		OutputDir: testDir,
		CacheDir:  testDir,
		TempDir:   testDir,
		Timeout:   10 * time.Second,
		MaxRetries: 1,
		FontPath:  "/System/Library/Fonts/Helvetica.ttc",
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
		
		// Verify file was created
		_, err = os.Stat(imagePath)
		assert.NoError(t, err, "image file should exist")
		assert.True(t, fileExists(imagePath), "image file should exist")
		
		// Verify file is not empty
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
		
		// Verify file was created
		_, err = os.Stat(imagePath)
		assert.NoError(t, err, "image file should exist")
		assert.True(t, fileExists(imagePath), "image file should exist")
		
		// Verify file is not empty
		fileInfo, _ := os.Stat(imagePath)
		assert.True(t, fileInfo.Size() > 0, "image file should not be empty")
	})
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}