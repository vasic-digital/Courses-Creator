package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/utils"
)

// CourseGenerator orchestrates the entire course generation process
type CourseGenerator struct {
	markdownParser *utils.MarkdownParser
	ttsProcessor   *TTSProcessor
	videoAssembler *VideoAssembler
	contentGen     *llm.CourseContentGenerator
	storage        *storage.StorageManager
}

// NewCourseGenerator creates a new course generator
func NewCourseGenerator() *CourseGenerator {
	// Create storage manager with default local storage
	storageConfigs := map[string]storage.StorageConfig{
		"default": storage.DefaultStorageConfig(),
	}
	
	storageManager, err := storage.NewStorageManager(storageConfigs)
	if err != nil {
		// Try to create a minimal storage manager with local directory
		config := storage.StorageConfig{
			Type:      "local",
			BasePath:  "./storage",
			PublicURL: "http://localhost:8080/storage",
		}
		storageManager, err = storage.NewStorageManagerWithDefault(config)
		if err != nil {
			// As a last resort, create a storage manager that uses temp directory
			tmpDir := os.TempDir()
			config = storage.StorageConfig{
				Type:      "local",
				BasePath:  filepath.Join(tmpDir, "course-creator-storage"),
				PublicURL: "http://localhost:8080/storage",
			}
			storageManager, _ = storage.NewStorageManagerWithDefault(config)
		}
	}
	
	return &CourseGenerator{
		markdownParser: utils.NewMarkdownParser(),
		ttsProcessor:   NewTTSProcessor(),
		videoAssembler: NewVideoAssembler(storageManager.DefaultProvider()),
		contentGen:     llm.NewCourseContentGenerator(),
		storage:        storageManager,
	}
}

// GenerateCourse generates a complete video course from markdown
func (cg *CourseGenerator) GenerateCourse(markdownPath, outputDir string, options models.ProcessingOptions) (*models.Course, error) {
	fmt.Printf("Starting course generation from %s\n", markdownPath)
	ctx := context.Background()

	// Validate inputs
	if markdownPath == "" {
		return nil, fmt.Errorf("markdown path cannot be empty")
	}
	if outputDir == "" {
		return nil, fmt.Errorf("output directory cannot be empty")
	}

	// Check if markdown file exists
	if _, err := os.Stat(markdownPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("markdown file does not exist: %s", markdownPath)
	}

	// Read markdown content
	content, err := readFile(markdownPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Parse markdown
	parsedCourse, err := cg.markdownParser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Generate enhanced course title and description using LLM
	title := parsedCourse.Title
	if title == "" {
		title, err = cg.contentGen.GenerateCourseTitle(ctx, content)
		if err != nil {
			fmt.Printf("Failed to generate title, using fallback: %v\n", err)
			title = "Untitled Course"
		}
	}

	description := parsedCourse.Description
	if description == "" {
		description, err = cg.contentGen.GenerateCourseDescription(ctx, title, content)
		if err != nil {
			fmt.Printf("Failed to generate description, using fallback: %v\n", err)
			description = "A comprehensive course on " + title
		}
	}

	// Generate metadata
	metadata, err := cg.contentGen.GenerateMetadata(ctx, title, description)
	if err != nil {
		fmt.Printf("Failed to generate enhanced metadata, using defaults: %v\n", err)
		metadata = map[string]interface{}{}
	}

	// Create course object
	course := &models.Course{
		ID:          fmt.Sprintf("course_%d", utils.HashString(content)),
		Title:       title,
		Description: description,
		Metadata: models.CourseMetadata{
			Author:   getStringFromMap(metadata, "author", getStringFromMap(parsedCourse.Metadata, "author", "Unknown")),
			Language: getStringFromMap(metadata, "language", getStringFromMap(parsedCourse.Metadata, "language", "en")),
			Tags:     getStringSliceFromMap(metadata, "tags", getStringSliceFromMap(parsedCourse.Metadata, "tags", []string{})),
		},
	}

	// Generate lessons with enhanced content
	var lessons []models.Lesson
	for _, section := range parsedCourse.Sections {
		lesson, err := cg.generateLesson(ctx, section, outputDir, options, content)
		if err != nil {
			return nil, fmt.Errorf("failed to generate lesson %s: %w", section.Title, err)
		}
		lessons = append(lessons, *lesson)
	}

	course.Lessons = lessons

	// Assemble final course
	err = cg.assembleCourse(course, outputDir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble course: %w", err)
	}

	return course, nil
}

// generateLesson generates a single lesson from section data
func (cg *CourseGenerator) generateLesson(ctx context.Context, section models.ParsedSection, outputDir string, options models.ProcessingOptions, courseContent string) (*models.Lesson, error) {
	fmt.Printf("Generating lesson: %s\n", section.Title)

	// Enhance lesson content using LLM
	enhancedContent, err := cg.contentGen.GenerateLessonContent(ctx, section.Title, section.Content)
	if err != nil {
		fmt.Printf("Failed to enhance lesson content, using original: %v\n", err)
		enhancedContent = section.Content
	}

	// Generate interactive elements
	interactiveElements, err := cg.contentGen.GenerateInteractiveElements(ctx, enhancedContent)
	if err != nil {
		fmt.Printf("Failed to generate interactive elements: %v\n", err)
		interactiveElements = []string{}
	}

	// Generate TTS audio
	audioPath, err := cg.ttsProcessor.GenerateAudio(enhancedContent, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audio: %w", err)
	}

	// Generate lesson ID
	lessonID := fmt.Sprintf("lesson_%d", utils.HashString(section.Content))
	
	// Create video - need course ID for storage
	courseID := fmt.Sprintf("course_%d", utils.HashString(courseContent))
	videoPath, err := cg.videoAssembler.CreateVideo(audioPath, enhancedContent, courseID, lessonID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	lesson := &models.Lesson{
		ID:                  lessonID,
		Title:               section.Title,
		Content:             enhancedContent,
		VideoURL:            &videoPath,
		AudioURL:            &audioPath,
		InteractiveElements:  parseInteractiveElements(interactiveElements),
		Order:               section.Order,
	}

	return lesson, nil
}

// assembleCourse assembles the final course package
func (cg *CourseGenerator) assembleCourse(course *models.Course, outputDir string, options models.ProcessingOptions) error {
	fmt.Println("Assembling final course package")

	// Placeholder for course assembly
	// - Generate course index
	// - Create player files
	// - Package everything

	return nil
}

// Helper functions
func readFile(path string) (string, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getStringSliceFromMap(m map[string]interface{}, key string, defaultValue []string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice
		}
		if slice, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return defaultValue
}

// parseInteractiveElements converts interactive element JSON strings to model objects
func parseInteractiveElements(elements []string) []models.InteractiveElement {
	var result []models.InteractiveElement
	for i, elementStr := range elements {
		// Simple parsing for now - in real implementation, parse JSON properly
		result = append(result, models.InteractiveElement{
			ID:       fmt.Sprintf("ie_%d", i),
			Type:     "quiz",
			Content:  elementStr,
			Position: i * 60, // Every 60 seconds
		})
	}
	return result
}
