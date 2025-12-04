package unit

import (
	"testing"

	"github.com/course-creator/core-processor/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Get underlying SQL DB to configure SQLite
	sqlDB, err := db.DB()
	require.NoError(t, err)
	// Enable foreign keys
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	// Configure SQLite for better concurrent access
	_, err = sqlDB.Exec("PRAGMA busy_timeout = 5000")
	require.NoError(t, err)
	_, err = sqlDB.Exec("PRAGMA journal_mode = WAL")
	require.NoError(t, err)
	
	// Migrate all tables
	err = db.AutoMigrate(
		&models.UserDB{},
		&models.UserPreferencesDB{},
		&models.UserSessionDB{},
		&models.CourseDB{},
		&models.JobDB{},
		&models.LessonDB{},
		&models.SubtitleDB{},
		&models.InteractiveElementDB{},
		&models.CourseMetadataDB{},
		&models.ProcessingJobDB{},
	)
	require.NoError(t, err)
	
	// Add BeforeCreate hooks for UUID generation
	db.Callback().Create().Before("gorm:create").Register("generate_user_id", func(db *gorm.DB) {
		if user, ok := db.Statement.Dest.(*models.UserDB); ok {
			if user.ID == "" {
				user.ID = uuid.New().String()
			}
		}
	})
	
	return db
}