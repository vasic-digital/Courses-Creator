package pipeline

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/course-creator/core-processor/config"
	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/llm"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/utils"
)

// CourseGenerator orchestrates the entire course generation process
type CourseGenerator struct {
	ttsProcessor     *TTSProcessor
	videoAssembler   *VideoAssembler
	diagramProcessor *DiagramProcessor
	contentGen       *llm.CourseContentGenerator
	storage          *storage.StorageManager
}

// NewCourseGenerator creates a new course generator
func NewCourseGenerator() *CourseGenerator {
	// Use default configuration
	cfg := &config.Config{
		Storage: map[string]config.StorageConfig{
			"default": config.StorageConfig{
				Type:      "local",
				BasePath:  "./storage",
				PublicURL: "http://localhost:8080/storage",
			},
		},
		TTS: config.TTSConfig{
			Provider: "bark",
			Timeout:  60 * time.Second,
		},
		LLM: config.LLMConfig{
			DefaultProvider:   "ollama",
			MaxCostPerRequest: 1.0,
			PrioritizeQuality: true,
			AllowPaid:         true,
			Ollama: config.OllamaConfig{
				BaseURL:      "http://localhost:11434",
				DefaultModel: "llama2",
			},
		},
	}

	factory := NewPipelineFactory(cfg)
	return factory.NewCourseGenerator()
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
	parsedCourse, err := utils.NewMarkdownParser().Parse(content)
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

	// Process diagrams in the content
	diagrams, err := cg.diagramProcessor.ProcessDiagrams(ctx, enhancedContent, options)
	if err != nil {
		fmt.Printf("Failed to process diagrams: %v\n", err)
		diagrams = []models.Diagram{} // Use empty slice as fallback
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
		Diagrams:            diagrams,
		InteractiveElements: parseInteractiveElements(interactiveElements),
		Order:               section.Order,
	}

	return lesson, nil
}

// assembleCourse assembles the final course package
func (cg *CourseGenerator) assembleCourse(course *models.Course, outputDir string, options models.ProcessingOptions) error {
	fmt.Println("Assembling final course package")

	// Create course index
	if err := cg.createCourseIndex(course, outputDir); err != nil {
		return fmt.Errorf("failed to create course index: %w", err)
	}

	// Create player configuration
	if err := cg.createPlayerConfig(course, outputDir); err != nil {
		return fmt.Errorf("failed to create player config: %w", err)
	}

	// Create course manifest
	if err := cg.createCourseManifest(course, outputDir); err != nil {
		return fmt.Errorf("failed to create course manifest: %w", err)
	}

	// Generate package metadata
	if err := cg.generatePackageMetadata(course, outputDir); err != nil {
		return fmt.Errorf("failed to generate package metadata: %w", err)
	}

	// If packaging is requested, create a distributable package
	if options.BackgroundMusic { // Using BackgroundMusic as package flag for now
		if err := cg.createCoursePackage(course, outputDir); err != nil {
			return fmt.Errorf("failed to create course package: %w", err)
		}
	}

	return nil
}

// createCourseIndex generates an HTML index file for the course
func (cg *CourseGenerator) createCourseIndex(course *models.Course, outputDir string) error {
	indexPath := filepath.Join(outputDir, "index.html")
	
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Course Index</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; line-height: 1.6; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .lessons { margin-top: 20px; }
        .lesson { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .lesson h3 { margin: 0 0 10px 0; color: #333; }
        .lesson-meta { color: #666; font-size: 0.9em; }
        .progress { margin-top: 20px; }
        .progress-bar { background: #e0e0e0; height: 20px; border-radius: 10px; overflow: hidden; }
        .progress-fill { background: #4CAF50; height: 100%%; transition: width 0.3s ease; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s</h1>
        <p>%s</p>
        <div class="progress">
            <h3>Progress: 0%% Complete</h3>
            <div class="progress-bar">
                <div class="progress-fill" style="width: 0%%"></div>
            </div>
        </div>
    </div>
    
    <div class="lessons">
        <h2>Lessons</h2>
`, course.Title, course.Description)

	for _, lesson := range course.Lessons {
		htmlContent += fmt.Sprintf(`
        <div class="lesson">
            <h3>Lesson %d: %s</h3>
            <p>Content for lesson %d</p>
            <div class="lesson-meta">
                Duration: %d seconds | Order: %d
            </div>
        </div>`, lesson.Order, lesson.Title, lesson.Order, lesson.Duration, lesson.Order)
	}

	htmlContent += `
    </div>
    
    <script>
        // Initialize course tracking
        var courseId = '` + course.ID + `';
        const lessons = document.querySelectorAll('.lesson');
        
        lessons.forEach((lesson, index) => {
            lesson.addEventListener('click', () => {
                // Navigate to lesson (implementation depends on player)
                console.log('Navigating to lesson ' + (index + 1));
            });
        });
        
        // Update progress from backend
        async function updateProgress() {
            try {
                const response = await fetch('/api/v1/courses/' + courseId + '/progress');
                const data = await response.json();
                document.querySelector('.progress-fill').style.width = data.progress + '%';
                document.querySelector('.progress h3').textContent = 'Progress: ' + data.progress + '% Complete';
            } catch (error) {
                console.error('Failed to update progress:', error);
            }
        }
        
        // Update progress periodically
        setInterval(updateProgress, 30000); // Every 30 seconds
    </script>
</body>
</html>`

	return os.WriteFile(indexPath, []byte(htmlContent), 0644)
}

// createPlayerConfig creates configuration files for the video player
func (cg *CourseGenerator) createPlayerConfig(course *models.Course, outputDir string) error {
	configPath := filepath.Join(outputDir, "player-config.json")
	
	config := map[string]interface{}{
		"courseId":     course.ID,
		"title":        course.Title,
		"description":  course.Description,
		"totalLessons": len(course.Lessons),
		"lessons": []map[string]interface{}{},
		"settings": map[string]interface{}{
			"autoplay":        true,
			"showControls":    true,
			"enableSubtitles": true,
			"theme":           "default",
		},
		"analytics": map[string]interface{}{
			"enabled": true,
			"trackProgress": true,
			"trackTimeSpent": true,
		},
	}

	for _, lesson := range course.Lessons {
		lessonData := map[string]interface{}{
			"id":          lesson.ID,
			"title":       lesson.Title,
			"order":       lesson.Order,
			"duration":    lesson.Duration,
		}
		
		if lesson.VideoURL != nil {
			lessonData["videoUrl"] = *lesson.VideoURL
		}
		if lesson.AudioURL != nil {
			lessonData["audioUrl"] = *lesson.AudioURL
		}
		if lesson.Content != "" {
			lessonData["content"] = lesson.Content
		}
		
		if len(lesson.Diagrams) > 0 {
			lessonData["diagrams"] = lesson.Diagrams
		}
		
		config["lessons"] = append(config["lessons"].([]map[string]interface{}), lessonData)
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal player config: %w", err)
	}

	return os.WriteFile(configPath, configJSON, 0644)
}

// createCourseManifest creates a manifest file with course metadata
func (cg *CourseGenerator) createCourseManifest(course *models.Course, outputDir string) error {
	manifestPath := filepath.Join(outputDir, "course-manifest.json")
	
	manifest := map[string]interface{}{
		"version":     "1.0",
		"generatedAt": time.Now().Format(time.RFC3339),
		"course": map[string]interface{}{
			"id":          course.ID,
			"title":       course.Title,
			"description": course.Description,
			"createdAt":   course.CreatedAt.Format(time.RFC3339),
			"updatedAt":   course.UpdatedAt.Format(time.RFC3339),
		},
		"assets": map[string]interface{}{
			"totalLessons":     len(course.Lessons),
			"hasVideo":         false,
			"hasAudio":         false,
			"hasDiagrams":      false,
			"totalDuration":    "0:00:00",
		},
		"compatibility": map[string]interface{}{
			"minPlayerVersion": "1.0.0",
			"platforms":        []string{"web", "desktop", "mobile"},
		},
	}

	// Check for assets
	var totalDurationSeconds int
	for _, lesson := range course.Lessons {
		if lesson.VideoURL != nil {
			manifest["assets"].(map[string]interface{})["hasVideo"] = true
		}
		if lesson.AudioURL != nil {
			manifest["assets"].(map[string]interface{})["hasAudio"] = true
		}
		if len(lesson.Diagrams) > 0 {
			manifest["assets"].(map[string]interface{})["hasDiagrams"] = true
		}
		
		// Calculate total duration from lesson durations in seconds
		var totalDurationSeconds int
		for _, lesson := range course.Lessons {
			if lesson.Duration > 0 {
				totalDurationSeconds += lesson.Duration
			}
			
			// Check for assets
			if lesson.VideoURL != nil {
				manifest["assets"].(map[string]interface{})["hasVideo"] = true
			}
			if lesson.AudioURL != nil {
				manifest["assets"].(map[string]interface{})["hasAudio"] = true
			}
			if len(lesson.Diagrams) > 0 {
				manifest["assets"].(map[string]interface{})["hasDiagrams"] = true
			}
		}
	}
	
	// Format total duration
	hours := totalDurationSeconds / 3600
	minutes := (totalDurationSeconds % 3600) / 60
	seconds := totalDurationSeconds % 60
	manifest["assets"].(map[string]interface{})["totalDuration"] = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal course manifest: %w", err)
	}

	return os.WriteFile(manifestPath, manifestJSON, 0644)
}

// generatePackageMetadata creates metadata for the course package
func (cg *CourseGenerator) generatePackageMetadata(course *models.Course, outputDir string) error {
	metadataPath := filepath.Join(outputDir, "metadata.json")
	
	metadata := map[string]interface{}{
		"packageType": "course",
		"version":     "1.0.0",
		"format":      "course-creator-v1",
		"generatedAt": time.Now().Format(time.RFC3339),
		"generator": map[string]interface{}{
			"name":    "Course Creator",
			"version": "1.0.0",
		},
		"course": course,
		"checksums": map[string]string{},
	}

	// Generate checksums for all files in output directory
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && path != metadataPath {
			relPath, _ := filepath.Rel(outputDir, path)
			hash := md5.Sum([]byte(path + info.ModTime().String())) // Simple checksum
			metadata["checksums"].(map[string]string)[relPath] = fmt.Sprintf("%x", hash)
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to generate checksums: %w", err)
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(metadataPath, metadataJSON, 0644)
}

// createCoursePackage creates a distributable package file
func (cg *CourseGenerator) createCoursePackage(course *models.Course, outputDir string) error {
	packagePath := filepath.Join(outputDir, fmt.Sprintf("%s.zip", course.ID))
	
	zipFile, err := os.Create(packagePath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add all files to the zip
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && filepath.Base(path) != fmt.Sprintf("%s.zip", course.ID) {
			relPath, _ := filepath.Rel(outputDir, path)
			
			fileInZip, err := zipWriter.Create(relPath)
			if err != nil {
				return fmt.Errorf("failed to create file in zip: %w", err)
			}
			
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file for zip: %w", err)
			}
			
			_, err = fileInZip.Write(fileContent)
			if err != nil {
				return fmt.Errorf("failed to write file to zip: %w", err)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to walk directory for packaging: %w", err)
	}

	fmt.Printf("Course package created: %s\n", packagePath)
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
