package database

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/course-creator/core-processor/models"
)

// DB represents the database connection
type DB struct {
	*gorm.DB
}

// Config holds database configuration
type Config struct {
	Path   string
	Debug  bool
	Env    string // "development", "production", "test"
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		Path:  "./data/course_creator.db",
		Debug: false,
		Env:   "development",
	}
}

// NewDatabase creates a new database connection
func NewDatabase(config *Config) (*DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Ensure data directory exists
	if err := os.MkdirAll(config.Path[:len(config.Path)-len("/course_creator.db")], 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Configure GORM logger
	var gormLogger logger.Interface
	if config.Debug {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(config.Path), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	// Auto-migrate schema
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{DB: db}, nil
}

// autoMigrate runs database migrations
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.UserDB{},
		&models.UserPreferencesDB{},
		&models.UserSessionDB{},
		&models.CourseDB{},
		&models.CourseMetadataDB{},
		&models.LessonDB{},
		&models.SubtitleDB{},
		&models.InteractiveElementDB{},
		&models.ProcessingJobDB{},
		&models.JobDB{},
	)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping checks if the database is accessible
func (db *DB) Ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	return sqlDB.Ping()
}

// GetGormDB returns the underlying GORM DB instance
func (db *DB) GetGormDB() *gorm.DB {
	return db.DB
}