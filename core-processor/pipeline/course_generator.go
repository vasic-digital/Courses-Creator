package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/utils"
)

// CourseGenerator orchestrates the entire course generation process
type CourseGenerator struct {
	markdownParser *utils.MarkdownParser
	ttsProcessor   *TTSProcessor
	videoAssembler *VideoAssembler
}

// NewCourseGenerator creates a new course generator
func NewCourseGenerator() *CourseGenerator {
	return &CourseGenerator{
		markdownParser: utils.NewMarkdownParser(),
		ttsProcessor:   NewTTSProcessor(),
		videoAssembler: NewVideoAssembler(),
	}
}

// GenerateCourse generates a complete video course from markdown
func (cg *CourseGenerator) GenerateCourse(markdownPath, outputDir string, options models.ProcessingOptions) (*models.Course, error) {
	fmt.Printf("Starting course generation from %s\n", markdownPath)

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

	// Create course object
	course := &models.Course{
		ID:          fmt.Sprintf("course_%s", filepath.Base(markdownPath)),
		Title:       parsedCourse.Title,
		Description: parsedCourse.Description,
		Metadata: models.CourseMetadata{
			Author:   getStringFromMap(parsedCourse.Metadata, "author", "Unknown"),
			Language: getStringFromMap(parsedCourse.Metadata, "language", "en"),
			Tags:     getStringSliceFromMap(parsedCourse.Metadata, "tags", []string{}),
		},
	}

	// Generate lessons
	var lessons []models.Lesson
	for _, section := range parsedCourse.Sections {
		lesson, err := cg.generateLesson(section, outputDir, options)
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
func (cg *CourseGenerator) generateLesson(section models.ParsedSection, outputDir string, options models.ProcessingOptions) (*models.Lesson, error) {
	fmt.Printf("Generating lesson: %s\n", section.Title)

	// Generate TTS audio
	audioPath, err := cg.ttsProcessor.GenerateAudio(section.Content, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audio: %w", err)
	}

	// Create video
	videoPath, err := cg.videoAssembler.CreateVideo(audioPath, section.Content, outputDir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	lesson := &models.Lesson{
		ID:       fmt.Sprintf("lesson_%d", hashString(section.Content)),
		Title:    section.Title,
		Content:  section.Content,
		VideoURL: &videoPath,
		AudioURL: &audioPath,
		Order:    section.Order,
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
	// Placeholder - implement file reading
	return "# Sample markdown content", nil
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
	}
	return defaultValue
}

func hashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}
