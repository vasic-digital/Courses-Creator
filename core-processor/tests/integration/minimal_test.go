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

func TestMinimalDatabase(t *testing.T) {
	// Setup in-memory database
	db, err := database.NewDatabase(&database.Config{
		Path:  ":memory:",
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	defer db.Close()

	// Create system user
	systemUser := &models.UserDB{
		ID:        "system",
		Email:     "system@example.com",
		Password:  "hashed_password",
		FirstName: "System",
		LastName:  "User",
		Role:      "admin",
		Active:    true,
	}
	err = db.GetGormDB().Create(systemUser).Error
	require.NoError(t, err, "Should create system user")

	// Get repository
	courseRepo := repository.NewCourseRepository(db)

	// Create a simple course
	courseID := uuid.New().String()
	course := &models.Course{
		ID:          courseID,
		Title:       "Minimal Test Course",
		Description: "Minimal course description",
	}

	t.Log("1. Creating course...")
	createdCourse, err := courseRepo.CreateCourse(course)
	if err != nil {
		t.Logf("CreateCourse error: %v", err)
	} else {
		t.Logf("CreateCourse succeeded")
	}
	require.NoError(t, err)
	require.NotNil(t, createdCourse)

	// Direct query 1
	t.Log("2. Direct query with db.GetGormDB()...")
	var directCourse models.CourseDB
	err = db.GetGormDB().First(&directCourse, "id = ?", courseID).Error
	if err != nil {
		t.Logf("Direct query error: %v", err)
	} else {
		t.Logf("Direct query found course: ID=%s", directCourse.ID)
	}
	assert.NoError(t, err)

	// GetCourseByID
	t.Log("3. Calling GetCourseByID...")
	retrievedCourse, err := courseRepo.GetCourseByID(courseID)
	if err != nil {
		t.Logf("GetCourseByID error: %v", err)
	} else {
		t.Logf("GetCourseByID succeeded: ID=%s", retrievedCourse.ID)
	}
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCourse)

	// GetAllCourses
	t.Log("4. Calling GetAllCourses...")
	courses, total, err := courseRepo.GetAllCourses(0, 10)
	if err != nil {
		t.Logf("GetAllCourses error: %v", err)
	} else {
		t.Logf("GetAllCourses returned %d courses, total=%d", len(courses), total)
	}
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(courses), 1, "Should have at least 1 course")
	assert.GreaterOrEqual(t, total, int64(1), "Total should be at least 1")
}
