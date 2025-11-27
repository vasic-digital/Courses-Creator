package pipeline

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/course-creator/core-processor/models"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/utils"
	"github.com/course-creator/core-processor/mcp_servers"
)

// DiagramProcessor handles diagram and illustration generation
type DiagramProcessor struct {
	config     DiagramConfig
	storage    storage.StorageInterface
	llava      *mcp_servers.LLaVAServer
}

// DiagramConfig holds diagram processing configuration
type DiagramConfig struct {
	Width       int
	Height      int
	Quality     int
	OutputDir   string
	CacheDir    string
	TempDir     string
	Timeout     time.Duration
	MaxRetries  int
	FontPath    string
}

// DiagramType represents different diagram types
type DiagramType string

const (
	DiagramFlowchart     DiagramType = "flowchart"
	DiagramSequence      DiagramType = "sequence"
	DiagramClass         DiagramType = "class"
	DiagramEntity        DiagramType = "entity"
	DiagramMindMap       DiagramType = "mindmap"
	DiagramArchitecture  DiagramType = "architecture"
	DiagramProcess       DiagramType = "process"
	DiagramConcept       DiagramType = "concept"
)

// Diagram represents a generated diagram
type Diagram struct {
	Type        DiagramType
	Title       string
	Description string
	ImagePath   string
	Data        map[string]interface{}
	Timestamp   time.Time
}

// DiagramRequest represents a diagram generation request
type DiagramRequest struct {
	Type        DiagramType
	Title       string
	Content     string
	Style       string
	Options     map[string]interface{}
}

// NewDiagramProcessor creates a new diagram processor
func NewDiagramProcessor(storage storage.StorageInterface) *DiagramProcessor {
	config := DiagramConfig{
		Width:      1920,
		Height:     1080,
		Quality:    90,
		OutputDir:  "/tmp/diagrams",
		CacheDir:   "/tmp/diagram_cache",
		TempDir:    "/tmp/diagram_temp",
		Timeout:    120 * time.Second,
		MaxRetries: 2,
		FontPath:   "/System/Library/Fonts/Helvetica.ttc", // macOS default
	}

	// Ensure directories exist
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.CacheDir)
	utils.EnsureDir(config.TempDir)

	dp := &DiagramProcessor{
		config:  config,
		storage: storage,
		llava:   mcp_servers.NewLLaVAServer(),
	}

	return dp
}

// NewDiagramProcessorWithConfig creates a diagram processor with custom config
func NewDiagramProcessorWithConfig(config DiagramConfig, storage storage.StorageInterface) *DiagramProcessor {
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.CacheDir)
	utils.EnsureDir(config.TempDir)

	return &DiagramProcessor{
		config:  config,
		storage: storage,
		llava:   mcp_servers.NewLLaVAServer(),
	}
}

// ProcessDiagrams detects and processes diagrams in content
func (dp *DiagramProcessor) ProcessDiagrams(ctx context.Context, content string, options models.ProcessingOptions) ([]models.Diagram, error) {
	fmt.Printf("Processing diagrams in content...\n")

	// Detect diagram markers in content
	diagramSections := dp.detectDiagrams(content)
	
	var diagrams []models.Diagram
	
	for i, section := range diagramSections {
		diagram, err := dp.generateDiagram(ctx, section, fmt.Sprintf("diagram_%d", i+1), options)
		if err != nil {
			fmt.Printf("Failed to generate diagram %d: %v\n", i+1, err)
			continue
		}
		
		diagrams = append(diagrams, *diagram)
	}

	fmt.Printf("Generated %d diagrams\n", len(diagrams))
	return diagrams, nil
}

// detectDiagrams identifies diagram sections in markdown content
func (dp *DiagramProcessor) detectDiagrams(content string) []DiagramRequest {
	var requests []DiagramRequest
	
	// Pattern for mermaid diagrams
	mermaidPattern := regexp.MustCompile("(?si)```mermaid\\s*\\n(.*?)\\n```")
	matches := mermaidPattern.FindAllStringSubmatch(content, -1)
	
	for i := range matches {
		if len(matches[i]) > 1 {
			diagramType, title := dp.analyzeMermaidDiagram(matches[i][1])
			
			requests = append(requests, DiagramRequest{
				Type:    diagramType,
				Title:   title,
				Content: matches[i][1],
				Style:   "mermaid",
				Options: map[string]interface{}{
					"format": "png",
					"theme":  "default",
				},
			})
		}
	}
	
	// Pattern for text-based diagram descriptions
	textDiagramPattern := regexp.MustCompile(`#{1,6}\s*(.*?[Dd]iagram.*?):\s*\n(.*?)`)
	textMatches := textDiagramPattern.FindAllStringSubmatch(content, -1)
	
	// Filter matches to ensure proper boundaries
	var filteredMatches [][]string
	for _, match := range textMatches {
		if len(match) > 2 {
			// Check if this is properly bounded by headers or document end
			startIndex := strings.Index(content, match[0])
			if startIndex >= 0 {
				// Look for next header or document end
				restOfContent := content[startIndex+len(match[0]):]
				nextHeaderIndex := strings.Index(restOfContent, "\n#")
				if nextHeaderIndex == -1 || nextHeaderIndex > 50 {
					filteredMatches = append(filteredMatches, match)
				}
			}
		}
	}
	
	for i := range textMatches {
		if len(textMatches[i]) > 2 {
			diagramType := dp.inferDiagramType(textMatches[i][1] + " " + textMatches[i][2])
			title := strings.TrimSpace(textMatches[i][1])
			
			requests = append(requests, DiagramRequest{
				Type:    diagramType,
				Title:   title,
				Content: strings.TrimSpace(textMatches[i][2]),
				Style:   "generated",
				Options: map[string]interface{}{
					"format": "png",
					"auto_layout": true,
				},
			})
		}
	}
	
	return requests
}

// analyzeMermaidDiagram determines diagram type from mermaid syntax
func (dp *DiagramProcessor) analyzeMermaidDiagram(content string) (DiagramType, string) {
	lines := strings.Split(content, "\n")
	title := "Untitled Diagram"
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
		}
		
		if strings.HasPrefix(line, "graph") || strings.HasPrefix(line, "flowchart") {
			return DiagramFlowchart, title
		} else if strings.HasPrefix(line, "sequenceDiagram") {
			return DiagramSequence, title
		} else if strings.HasPrefix(line, "classDiagram") {
			return DiagramClass, title
		} else if strings.HasPrefix(line, "erDiagram") {
			return DiagramEntity, title
		} else if strings.HasPrefix(line, "mindmap") {
			return DiagramMindMap, title
		} else if strings.Contains(line, "graph TB") || strings.Contains(line, "graph TD") {
			return DiagramArchitecture, title
		}
	}
	
	return DiagramConcept, title
}

// inferDiagramType infers diagram type from text content
func (dp *DiagramProcessor) inferDiagramType(content string) DiagramType {
	content = strings.ToLower(content)
	
	if strings.Contains(content, "flow") || strings.Contains(content, "process") || strings.Contains(content, "step") {
		return DiagramFlowchart
	} else if strings.Contains(content, "sequence") || strings.Contains(content, "message") || strings.Contains(content, "interaction") {
		return DiagramSequence
	} else if strings.Contains(content, "class") || strings.Contains(content, "inheritance") || strings.Contains(content, "method") {
		return DiagramClass
	} else if strings.Contains(content, "entity") || strings.Contains(content, "relationship") || strings.Contains(content, "database") {
		return DiagramEntity
	} else if strings.Contains(content, "mind") || strings.Contains(content, "idea") || strings.Contains(content, "branch") {
		return DiagramMindMap
	} else if strings.Contains(content, "architecture") || strings.Contains(content, "system") || strings.Contains(content, "component") {
		return DiagramArchitecture
	} else if strings.Contains(content, "concept") || strings.Contains(content, "idea") || strings.Contains(content, "model") {
		return DiagramConcept
	}
	
	return DiagramProcess // Default
}

// generateDiagram generates a diagram from request
func (dp *DiagramProcessor) generateDiagram(ctx context.Context, request DiagramRequest, filename string, options models.ProcessingOptions) (*models.Diagram, error) {
	fmt.Printf("Generating %s diagram: %s\n", request.Type, request.Title)
	
	var imagePath string
	var err error
	
	switch request.Style {
	case "mermaid":
		imagePath, err = dp.generateMermaidDiagram(ctx, request, filename)
	case "generated":
		imagePath, err = dp.generateTextBasedDiagram(ctx, request, filename)
	default:
		imagePath, err = dp.generateDefaultDiagram(ctx, request, filename)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to generate diagram: %w", err)
	}
	
	// Save to storage
	diagramStoragePath := dp.getStoragePath(request, filename)
	
	// Read generated image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read diagram image: %w", err)
	}
	
	// Save to storage
	err = dp.storage.Save(diagramStoragePath, imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to save diagram to storage: %w", err)
	}
	
	// Clean up temporary file
	os.Remove(imagePath)
	
	diagram := &models.Diagram{
		ID:          fmt.Sprintf("diagram_%d", utils.HashString(request.Content+request.Title)),
		Type:        string(request.Type),
		Title:       request.Title,
		Description: dp.generateDescription(request),
		ImageURL:    &diagramStoragePath,
		Data:        request.Options,
		CreatedAt:   time.Now(),
	}
	
	return diagram, nil
}

// generateMermaidDiagram generates diagram from mermaid syntax
func (dp *DiagramProcessor) generateMermaidDiagram(ctx context.Context, request DiagramRequest, filename string) (string, error) {
	outputPath := filepath.Join(dp.config.OutputDir, filename+".png")
	
	// For now, create a placeholder implementation
	// In production, this would use mermaid-cli or similar
	return dp.createPlaceholderDiagram(outputPath, request, "Mermaid")
}

// generateTextBasedDiagram generates diagram from text description
func (dp *DiagramProcessor) generateTextBasedDiagram(ctx context.Context, request DiagramRequest, filename string) (string, error) {
	outputPath := filepath.Join(dp.config.OutputDir, filename+".png")
	
	// Create diagram based on type and content
	return dp.createTypedDiagram(outputPath, request)
}

// generateDefaultDiagram creates a simple default diagram
func (dp *DiagramProcessor) generateDefaultDiagram(ctx context.Context, request DiagramRequest, filename string) (string, error) {
	outputPath := filepath.Join(dp.config.OutputDir, filename+".png")
	return dp.createPlaceholderDiagram(outputPath, request, "Default")
}

// createPlaceholderDiagram creates a placeholder diagram image
func (dp *DiagramProcessor) createPlaceholderDiagram(outputPath string, request DiagramRequest, generator string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, dp.config.Width, dp.config.Height))
	
	// Background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 240, G: 240, B: 240, A: 255}}, image.Point{}, draw.Src)
	
	// Add title text (simplified - would use proper text rendering in production)
	titleColor := color.RGBA{R: 50, G: 50, B: 50, A: 255}
	
	// Create simple shapes based on diagram type
	switch request.Type {
	case DiagramFlowchart:
		dp.drawFlowchartShapes(img, titleColor)
	case DiagramSequence:
		dp.drawSequenceShapes(img, titleColor)
	case DiagramClass:
		dp.drawClassShapes(img, titleColor)
	case DiagramMindMap:
		dp.drawMindMapShapes(img, titleColor)
	default:
		dp.drawGenericShapes(img, titleColor)
	}
	
	// Save image
	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	err = png.Encode(file, img)
	if err != nil {
		return "", err
	}
	
	return outputPath, nil
}

// createTypedDiagram creates diagram based on type and content
func (dp *DiagramProcessor) createTypedDiagram(outputPath string, request DiagramRequest) (string, error) {
	// Analyze content and create appropriate diagram
	content := strings.ToLower(request.Content)
	
	// Count key elements
	elementCount := dp.countElements(content)
	
	img := image.NewRGBA(image.Rect(0, 0, dp.config.Width, dp.config.Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 250, G: 250, B: 250, A: 255}}, image.Point{}, draw.Src)
	
	// Generate diagram based on type and content
	switch request.Type {
	case DiagramFlowchart:
		dp.generateFlowchartFromContent(img, content, elementCount)
	case DiagramMindMap:
		dp.generateMindMapFromContent(img, content, elementCount)
	default:
		dp.generateGenericFromContent(img, content, elementCount)
	}
	
	// Save image
	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	err = png.Encode(file, img)
	if err != nil {
		return "", err
	}
	
	return outputPath, nil
}

// countElements estimates number of elements in content
func (dp *DiagramProcessor) countElements(content string) int {
	// Simple heuristic to count elements
	elements := 0
	
	// Look for bullet points, numbers, and keywords
	bullets := regexp.MustCompile(`[-*+]\s+`)
	elements += len(bullets.FindAllString(content, -1))
	
	numbers := regexp.MustCompile(`\d+\.\s+`)
	elements += len(numbers.FindAllString(content, -1))
	
	// Look for keywords
	keywords := []string{"step", "stage", "phase", "process", "item", "element", "node", "box"}
	for _, keyword := range keywords {
		elements += strings.Count(content, keyword)
	}
	
	if elements == 0 {
		elements = 3 // Minimum for a meaningful diagram
	}
	
	return elements
}

// Drawing methods for different diagram types
func (dp *DiagramProcessor) drawFlowchartShapes(img *image.RGBA, titleColor color.RGBA) {
	// Draw flowchart shapes (rectangles and arrows)
	positions := [][]int{{200, 100}, {200, 300}, {200, 500}, {200, 700}}
	
	for i, pos := range positions {
		dp.drawRectangle(img, pos[0], pos[1], 300, 80, dp.getColorForIndex(i))
		if i < len(positions)-1 {
			nextPos := positions[i+1]
			dp.drawArrow(img, pos[0]+150, pos[1]+80, nextPos[0]+150, nextPos[1], dp.getColorForIndex(i))
		}
	}
}

func (dp *DiagramProcessor) drawSequenceShapes(img *image.RGBA, titleColor color.RGBA) {
	// Draw sequence diagram (boxes and vertical lines)
	boxes := [][]int{{100, 100}, {500, 100}, {900, 100}, {1300, 100}}
	
	for _, box := range boxes {
		dp.drawRectangle(img, box[0], box[1], 120, 60, titleColor)
		dp.drawLine(img, box[0]+60, box[1]+60, box[0]+60, 800, titleColor)
	}
}

func (dp *DiagramProcessor) drawClassShapes(img *image.RGBA, titleColor color.RGBA) {
	// Draw class diagram (rectangles with compartments)
	positions := [][]int{{200, 100}, {600, 100}, {1000, 100}}
	
	for _, pos := range positions {
		// Main box
		dp.drawRectangle(img, pos[0], pos[1], 280, 200, titleColor)
		// Title compartment
		dp.drawLine(img, pos[0], pos[1]+40, pos[0]+280, pos[1]+40, titleColor)
		// Methods compartment
		dp.drawLine(img, pos[0], pos[1]+120, pos[0]+280, pos[1]+120, titleColor)
	}
}

func (dp *DiagramProcessor) drawMindMapShapes(img *image.RGBA, titleColor color.RGBA) {
	// Draw mind map (central node with branches)
	centerX, centerY := 960, 400
	branches := 8
	
	for i := 0; i < branches; i++ {
		angle := float64(i) * 2 * math.Pi / float64(branches)
		endX := centerX + int(300*math.Cos(angle))
		endY := centerY + int(300*math.Sin(angle))
		
		dp.drawCircle(img, centerX, centerY, 40, dp.getColorForIndex(0))
		dp.drawLine(img, centerX, centerY, endX, endY, titleColor)
		dp.drawCircle(img, endX, endY, 25, dp.getColorForIndex(i))
	}
}

func (dp *DiagramProcessor) drawGenericShapes(img *image.RGBA, titleColor color.RGBA) {
	// Draw generic shapes
	for i := 0; i < 4; i++ {
		x := 200 + (i % 2) * 600
		y := 200 + (i / 2) * 300
		dp.drawRectangle(img, x, y, 200, 150, dp.getColorForIndex(i))
	}
}

// Helper drawing methods
func (dp *DiagramProcessor) drawRectangle(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if dx >= 0 && dy >= 0 && x+dx < dp.config.Width && y+dy < dp.config.Height {
				img.Set(x+dx, y+dy, c)
			}
		}
	}
}

func (dp *DiagramProcessor) drawCircle(img *image.RGBA, centerX, centerY, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				x, y := centerX+dx, centerY+dy
				if x >= 0 && y >= 0 && x < dp.config.Width && y < dp.config.Height {
					img.Set(x, y, c)
				}
			}
		}
	}
}

func (dp *DiagramProcessor) drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	
	for {
		if x1 >= 0 && y1 >= 0 && x1 < dp.config.Width && y1 < dp.config.Height {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (dp *DiagramProcessor) drawArrow(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dp.drawLine(img, x1, y1, x2, y2, c)
	
	// Draw arrowhead
	angle := math.Atan2(float64(y2-y1), float64(x2-x1))
	arrowLength := float64(10)
	arrowAngle := 0.5
	
	// Calculate arrowhead points
	px := float64(x2) - arrowLength*math.Cos(angle-arrowAngle)
	py := float64(y2) - arrowLength*math.Sin(angle-arrowAngle)
	qx := float64(x2) - arrowLength*math.Cos(angle+arrowAngle)
	qy := float64(y2) - arrowLength*math.Sin(angle+arrowAngle)
	
	dp.drawLine(img, x2, y2, int(px), int(py), c)
	dp.drawLine(img, x2, y2, int(qx), int(qy), c)
}

func (dp *DiagramProcessor) getColorForIndex(index int) color.RGBA {
	colors := []color.RGBA{
		{R: 74, G: 144, B: 226, A: 255},   // Blue
		{R: 80, G: 200, B: 120, A: 255},   // Green
		{R: 245, G: 166, B: 35, A: 255},   // Orange
		{R: 155, G: 89, B: 182, A: 255},   // Purple
		{R: 231, G: 76, B: 60, A: 255},    // Red
		{R: 52, G: 152, B: 219, A: 255},   // Light Blue
	}
	
	return colors[index%len(colors)]
}

func (dp *DiagramProcessor) generateFlowchartFromContent(img *image.RGBA, content string, elementCount int) {
	// Simple flowchart generation based on content
	fmt.Printf("Generating flowchart with %d elements\n", elementCount)
	
	// Create vertical flow
	elementsPerColumn := 5
	cols := (elementCount + elementsPerColumn - 1) / elementsPerColumn
	
	for i := 0; i < elementCount; i++ {
		col := i / elementsPerColumn
		row := i % elementsPerColumn
		
		x := 200 + col*400
		y := 100 + row*120
		
		dp.drawRectangle(img, x, y, 300, 80, dp.getColorForIndex(i))
		
		// Draw arrows
		if row < elementsPerColumn-1 && i+1 < elementCount {
			nextY := 100 + (row+1)*120
			dp.drawArrow(img, x+150, y+80, x+150, nextY, dp.getColorForIndex(0))
		} else if row == elementsPerColumn-1 && col < cols-1 && i+1 < elementCount {
			nextX := 200 + (col+1)*400
			dp.drawArrow(img, x+150, y+40, nextX, y+40, dp.getColorForIndex(0))
		}
	}
}

func (dp *DiagramProcessor) generateMindMapFromContent(img *image.RGBA, content string, elementCount int) {
	// Simple mind map generation
	centerX, centerY := 960, 400
	
	dp.drawCircle(img, centerX, centerY, 60, dp.getColorForIndex(0))
	
	branches := min(elementCount-1, 12)
	for i := 0; i < branches; i++ {
		angle := float64(i) * 2 * math.Pi / float64(branches)
		radius := float64(300 + (i%3)*50) // Vary radius for visual interest
		endX := centerX + int(radius*math.Cos(angle))
		endY := centerY + int(radius*math.Sin(angle))
		
		dp.drawLine(img, centerX, centerY, endX, endY, dp.getColorForIndex(0))
		dp.drawCircle(img, endX, endY, 30, dp.getColorForIndex(i+1))
	}
}

func (dp *DiagramProcessor) generateGenericFromContent(img *image.RGBA, content string, elementCount int) {
	// Generic diagram layout
	gridSize := 4
	cols := (elementCount + gridSize - 1) / gridSize
	
	for i := 0; i < elementCount; i++ {
		col := i % cols
		row := i / cols
		
		x := 100 + col*450
		y := 100 + row*200
		
		dp.drawRectangle(img, x, y, 400, 150, dp.getColorForIndex(i))
	}
}

func (dp *DiagramProcessor) getStoragePath(request DiagramRequest, filename string) string {
	return fmt.Sprintf("diagrams/%s/%s.png", request.Type, filename)
}

func (dp *DiagramProcessor) generateDescription(request DiagramRequest) string {
	return fmt.Sprintf("A %s diagram showing %s", request.Type, request.Title)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}