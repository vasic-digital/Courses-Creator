package integration

import (
	"testing"

	"github.com/course-creator/core-processor/database"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugDatabase(t *testing.T) {
	// Setup in-memory database
	db, err := database.NewDatabase(&database.Config{
		Path:  ":memory:",
		Debug: true, // Enable debug logging
		Env:   "test",
	})
	require.NoError(t, err)
	defer db.Close()

	// Get repository
	courseRepo := repository.NewCourseRepository(db)

	// Create a simple course
	course := &models.Course{
		ID:          uuid.New().String(),
		Title:       "Debug Test Course",
		Description: "Debug course description",
	}

	t.Log("Creating course...")
	createdCourse, err := courseRepo.CreateCourse(course)
	if err != nil {
		t.Logf("CreateCourse returned error: %v", err)
	} else {
		t.Logf("CreateCourse succeeded, createdCourse: %v", createdCourse)
	}

	// Try to get the course directly
	t.Log("Trying to get course directly from database...")
	var courseDB models.CourseDB
	err = db.First(&courseDB, "id = ?", course.ID).Error
	if err != nil {
		t.Logf("Direct database query error: %v", err)
	} else {
		t.Logf("Direct database query found course: %v", courseDB)
	}

	// Count all courses
	var count int64
	err = db.Model(&models.CourseDB{}).Count(&count).Error
	if err != nil {
		t.Logf("Count error: %v", err)
	} else {
		t.Logf("Total courses in database: %d", count)
	}

	// List all courses
	var courses []models.CourseDB
	err = db.Find(&courses).Error
	if err != nil {
		t.Logf("Find all error: %v", err)
	} else {
		t.Logf("Found %d courses:", len(courses))
		for i, c := range courses {
			t.Logf("  Course %d: ID=%s, Title=%s, UserID=%s", i+1, c.ID, c.Title, c.UserID)
		}
	}

	assert.NoError(t, err, "Database operations should not error")
}
