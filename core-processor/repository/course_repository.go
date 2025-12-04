package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
)

// CourseRepository handles course data operations
type CourseRepository struct {
	db *database.DB
}

// NewCourseRepository creates a new course repository
func NewCourseRepository(db *database.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// CreateCourse creates a new course
func (r *CourseRepository) CreateCourse(course *models.Course) (*models.CourseDB, error) {
	courseDB := &models.CourseDB{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Start transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create course
	if err := tx.Create(courseDB).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	// Create metadata
	if course.Metadata.Author != "" || course.Metadata.Language != "" {
		tagsJSON, _ := json.Marshal(course.Metadata.Tags)
		metadataDB := &models.CourseMetadataDB{
			CourseID:      courseDB.ID,
			Author:        course.Metadata.Author,
			Language:      course.Metadata.Language,
			Tags:          string(tagsJSON),
			ThumbnailURL:  course.Metadata.ThumbnailURL,
			TotalDuration: course.Metadata.TotalDuration,
		}
		if err := tx.Create(metadataDB).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create course metadata: %w", err)
		}
		courseDB.Metadata = *metadataDB
	}

	// Create lessons
	for _, lesson := range course.Lessons {
		lessonDB := &models.LessonDB{
			ID:       lesson.ID,
			CourseID: courseDB.ID,
			Title:    lesson.Title,
			Content:  lesson.Content,
			VideoURL: lesson.VideoURL,
			AudioURL: lesson.AudioURL,
			Duration: lesson.Duration,
			Order:    lesson.Order,
		}
		if err := tx.Create(lessonDB).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create lesson: %w", err)
		}

		// Create subtitles
		for _, subtitle := range lesson.Subtitles {
			timestampsJSON, _ := json.Marshal(subtitle.Timestamps)
			subtitleDB := &models.SubtitleDB{
				LessonID:   lessonDB.ID,
				Language:   subtitle.Language,
				Content:    subtitle.Content,
				Timestamps: string(timestampsJSON),
			}
			if err := tx.Create(subtitleDB).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create subtitle: %w", err)
			}
		}

		// Create interactive elements
		for _, element := range lesson.InteractiveElements {
			elementDB := &models.InteractiveElementDB{
				LessonID: lessonDB.ID,
				ID:       element.ID,
				Type:     element.Type,
				Content:  element.Content,
				Position: element.Position,
			}
			if err := tx.Create(elementDB).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create interactive element: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load the complete course with relations
	return r.GetCourseByID(courseDB.ID)
}

// GetCourseByID retrieves a course by ID
func (r *CourseRepository) GetCourseByID(id string) (*models.CourseDB, error) {
	var course models.CourseDB
	if err := r.db.Preload("Metadata").Preload("Lessons.Subtitles").Preload("Lessons.InteractiveElements").First(&course, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("course not found")
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	return &course, nil
}

// GetAllCourses retrieves all courses with pagination
func (r *CourseRepository) GetAllCourses(offset, limit int) ([]models.CourseDB, int64, error) {
	var courses []models.CourseDB
	var total int64

	// Count total courses
	if err := r.db.Model(&models.CourseDB{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %w", err)
	}

	// Get courses with pagination
	if err := r.db.Preload("Metadata").Offset(offset).Limit(limit).Find(&courses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get courses: %w", err)
	}

	return courses, total, nil
}

// UpdateCourse updates an existing course
func (r *CourseRepository) UpdateCourse(id string, course *models.Course) (*models.CourseDB, error) {
	existing, err := r.GetCourseByID(id)
	if err != nil {
		return nil, err
	}

	// Update course fields
	existing.Title = course.Title
	existing.Description = course.Description
	existing.UpdatedAt = time.Now()

	// Start transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update course
	if err := tx.Save(existing).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update course: %w", err)
	}

	// Update metadata if exists
	if course.Metadata.Author != "" || course.Metadata.Language != "" {
		tagsJSON, _ := json.Marshal(course.Metadata.Tags)
		if existing.Metadata.CourseID == "" {
			// Create metadata if doesn't exist
			metadataDB := &models.CourseMetadataDB{
				CourseID:      existing.ID,
				Author:        course.Metadata.Author,
				Language:      course.Metadata.Language,
				Tags:          string(tagsJSON),
				ThumbnailURL:  course.Metadata.ThumbnailURL,
				TotalDuration: course.Metadata.TotalDuration,
			}
			if err := tx.Create(metadataDB).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create course metadata: %w", err)
			}
		} else {
			// Update existing metadata
			if err := tx.Model(&existing.Metadata).Updates(map[string]interface{}{
				"author":         course.Metadata.Author,
				"language":       course.Metadata.Language,
				"tags":           string(tagsJSON),
				"thumbnail_url":  course.Metadata.ThumbnailURL,
				"total_duration": course.Metadata.TotalDuration,
				"updated_at":     time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update course metadata: %w", err)
			}
		}
	}

	// Update lessons
	if len(course.Lessons) > 0 {
		// Delete existing lessons and related data
		if err := tx.Where("course_id = ?", existing.ID).Delete(&models.LessonDB{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete existing lessons: %w", err)
		}

		// Create new lessons
		for _, lesson := range course.Lessons {
			lessonDB := &models.LessonDB{
				ID:       lesson.ID,
				CourseID: existing.ID,
				Title:    lesson.Title,
				Content:  lesson.Content,
				VideoURL: lesson.VideoURL,
				AudioURL: lesson.AudioURL,
				Duration: lesson.Duration,
				Order:    lesson.Order,
			}
			if err := tx.Create(lessonDB).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create lesson: %w", err)
			}

			// Create subtitles and interactive elements (same as in CreateCourse)
			for _, subtitle := range lesson.Subtitles {
				timestampsJSON, _ := json.Marshal(subtitle.Timestamps)
				subtitleDB := &models.SubtitleDB{
					LessonID:   lessonDB.ID,
					Language:   subtitle.Language,
					Content:    subtitle.Content,
					Timestamps: string(timestampsJSON),
				}
				if err := tx.Create(subtitleDB).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to create subtitle: %w", err)
				}
			}

			for _, element := range lesson.InteractiveElements {
				elementDB := &models.InteractiveElementDB{
					LessonID: lessonDB.ID,
					ID:       element.ID,
					Type:     element.Type,
					Content:  element.Content,
					Position: element.Position,
				}
				if err := tx.Create(elementDB).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to create interactive element: %w", err)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetCourseByID(id)
}

// DeleteCourse deletes a course by ID
func (r *CourseRepository) DeleteCourse(id string) error {
	if err := r.db.Delete(&models.CourseDB{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	return nil
}

// SearchCourses searches courses by title or description
func (r *CourseRepository) SearchCourses(query string, offset, limit int) ([]models.CourseDB, int64, error) {
	var courses []models.CourseDB
	var total int64

	searchPattern := "%" + query + "%"

	// Count matching courses
	if err := r.db.Model(&models.CourseDB{}).Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %w", err)
	}

	// Get courses with pagination
	if err := r.db.Preload("Metadata").Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern).Offset(offset).Limit(limit).Find(&courses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search courses: %w", err)
	}

	return courses, total, nil
}
