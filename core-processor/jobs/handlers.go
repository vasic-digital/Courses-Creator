package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	filestorage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/metrics"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/pipeline"
	"github.com/course-creator/core-processor/utils"
)

// JobContext provides context for job handlers
type JobContext struct {
	Queue           *JobQueue
	Storage         filestorage.StorageInterface
	MarkdownParser  *utils.MarkdownParser
	CourseGenerator *pipeline.CourseGenerator
}

// RegisterDefaultHandlers registers default job handlers
func (jc *JobContext) RegisterDefaultHandlers() {
	jc.Queue.RegisterHandler(JobTypeCourseGeneration, jc.HandleCourseGeneration)
	jc.Queue.RegisterHandler(JobTypeVideoProcessing, jc.HandleVideoProcessing)
	jc.Queue.RegisterHandler(JobTypeAudioGeneration, jc.HandleAudioGeneration)
	jc.Queue.RegisterHandler(JobTypeSubtitleGeneration, jc.HandleSubtitleGeneration)
}

// HandleCourseGeneration handles course generation jobs
func (jc *JobContext) HandleCourseGeneration(ctx context.Context, job *Job) error {
	log.Printf("Starting course generation for job %s", job.ID)
	startTime := time.Now()

	// Extract job parameters
	inputPath, ok := job.Payload["input_path"].(string)
	if !ok {
		return fmt.Errorf("input_path is required")
	}

	outputPath, ok := job.Payload["output_path"].(string)
	if !ok {
		return fmt.Errorf("output_path is required")
	}

	// Extract processing options
	var options *models.ProcessingOptions
	if opts, ok := job.Payload["options"]; ok {
		if optsMap, ok := opts.(map[string]interface{}); ok {
			options = &models.ProcessingOptions{
				BackgroundMusic: true,           // Default
				Quality:         "standard",     // Default
				Languages:       []string{"en"}, // Default
			}

			if voice, ok := optsMap["voice"].(string); ok {
				options.Voice = &voice
			}
			if bgMusic, ok := optsMap["background_music"].(bool); ok {
				options.BackgroundMusic = bgMusic
			}
			if bgStyle, ok := optsMap["background_style"].(string); ok {
				options.BackgroundStyle = bgStyle
			}
			if langs, ok := optsMap["languages"].([]interface{}); ok {
				var languages []string
				for _, lang := range langs {
					if str, ok := lang.(string); ok {
						languages = append(languages, str)
					}
				}
				if len(languages) > 0 {
					options.Languages = languages
				}
			}
			if quality, ok := optsMap["quality"].(string); ok {
				options.Quality = quality
			}
		}
	} else {
		// Use default options
		options = &models.ProcessingOptions{
			BackgroundMusic: true,
			Quality:         "standard",
			Languages:       []string{"en"},
		}
	}

	// Update progress
	jc.Queue.UpdateProgress(job.ID, 5)

	// Generate course
	result, err := jc.CourseGenerator.GenerateCourse(inputPath, outputPath, *options)
	if err != nil {
		metrics.RecordCourseGeneration("failed", options.Quality, time.Since(startTime))
		return fmt.Errorf("failed to generate course: %w", err)
	}

	// Record success metrics
	metrics.RecordCourseGeneration("completed", options.Quality, time.Since(startTime))

	// Store result
	jc.Queue.UpdateResult(job.ID, map[string]interface{}{
		"course_id":    result.ID,
		"output_path":  outputPath,
		"duration":     300, // Example duration
		"lesson_count": len(result.Lessons),
	})

	log.Printf("Course generation completed for job %s", job.ID)
	return nil
}

// HandleVideoProcessing handles video processing jobs
func (jc *JobContext) HandleVideoProcessing(ctx context.Context, job *Job) error {
	log.Printf("Starting video processing for job %s", job.ID)

	// Extract job parameters
	courseID, ok := job.Payload["course_id"].(string)
	if !ok {
		return fmt.Errorf("course_id is required")
	}

	lessonID, ok := job.Payload["lesson_id"].(string)
	if !ok {
		return fmt.Errorf("lesson_id is required")
	}

	// Update progress
	jc.Queue.UpdateProgress(job.ID, 10)

	// Get lesson from database or storage
	// This would typically fetch the lesson content and process it

	// Simulate video processing steps
	steps := []string{
		"Preparing audio for video assembly",
		"Generating background visuals",
		"Creating text overlays",
		"Assembling video components",
		"Applying post-processing effects",
		"Finalizing video output",
	}

	for i, step := range steps {
		log.Printf("Video processing step %d/%d: %s", i+1, len(steps), step)
		time.Sleep(1 * time.Second) // Simulate processing time

		// Update progress (10% initial + 90% distributed across steps)
		progress := 10 + (90 * (i + 1) / len(steps))
		jc.Queue.UpdateProgress(job.ID, progress)
	}

	// Store result
	jc.Queue.UpdateResult(job.ID, map[string]interface{}{
		"course_id": courseID,
		"lesson_id": lessonID,
		"video_url": fmt.Sprintf("/storage/videos/%s/%s.mp4", courseID, lessonID),
		"duration":  300, // Example duration in seconds
	})

	log.Printf("Video processing completed for job %s", job.ID)
	return nil
}

// HandleAudioGeneration handles audio generation jobs
func (jc *JobContext) HandleAudioGeneration(ctx context.Context, job *Job) error {
	log.Printf("Starting audio generation for job %s", job.ID)

	// Extract job parameters
	text, ok := job.Payload["text"].(string)
	if !ok {
		return fmt.Errorf("text is required")
	}

	voice, ok := job.Payload["voice"].(string)
	if !ok {
		voice = "default" // Default voice
	}

	// Update progress
	jc.Queue.UpdateProgress(job.ID, 10)

	// Simulate audio generation steps
	steps := []string{
		"Preprocessing text for TTS",
		"Generating speech using TTS engine",
		"Applying audio processing",
		"Normalizing audio levels",
		"Adding background music if requested",
		"Finalizing audio output",
	}

	for i, step := range steps {
		log.Printf("Audio generation step %d/%d: %s", i+1, len(steps), step)
		time.Sleep(800 * time.Millisecond) // Simulate processing time

		// Update progress (10% initial + 90% distributed across steps)
		progress := 10 + (90 * (i + 1) / len(steps))
		jc.Queue.UpdateProgress(job.ID, progress)
	}

	// Generate audio file path
	audioID := job.ID
	audioURL := fmt.Sprintf("/storage/audio/%s.mp3", audioID)

	// Store result
	jc.Queue.UpdateResult(job.ID, map[string]interface{}{
		"audio_url": audioURL,
		"duration":  len(text) / 10, // Rough estimate: 10 characters per second
		"voice":     voice,
	})

	log.Printf("Audio generation completed for job %s", job.ID)
	return nil
}

// HandleSubtitleGeneration handles subtitle generation jobs
func (jc *JobContext) HandleSubtitleGeneration(ctx context.Context, job *Job) error {
	log.Printf("Starting subtitle generation for job %s", job.ID)

	// Extract job parameters
	audioURL, ok := job.Payload["audio_url"].(string)
	if !ok {
		return fmt.Errorf("audio_url is required")
	}

	language, ok := job.Payload["language"].(string)
	if !ok {
		language = "en" // Default language
	}

	// Update progress
	jc.Queue.UpdateProgress(job.ID, 10)

	// Simulate subtitle generation steps
	steps := []string{
		"Extracting audio features",
		"Performing speech recognition",
		"Generating time-coded text",
		"Optimizing subtitle timing",
		"Formatting subtitle output",
		"Validating subtitle synchronization",
	}

	for i, step := range steps {
		log.Printf("Subtitle generation step %d/%d: %s", i+1, len(steps), step)
		time.Sleep(600 * time.Millisecond) // Simulate processing time

		// Update progress (10% initial + 90% distributed across steps)
		progress := 10 + (90 * (i + 1) / len(steps))
		jc.Queue.UpdateProgress(job.ID, progress)
	}

	// Generate subtitle file path
	subtitleID := job.ID
	subtitleURL := fmt.Sprintf("/storage/subtitles/%s_%s.srt", subtitleID, language)

	// Store result
	jc.Queue.UpdateResult(job.ID, map[string]interface{}{
		"subtitle_url": subtitleURL,
		"language":     language,
		"audio_url":    audioURL,
	})

	log.Printf("Subtitle generation completed for job %s", job.ID)
	return nil
}
