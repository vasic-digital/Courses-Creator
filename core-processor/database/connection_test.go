package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	require.NotNil(t, config)
	assert.Equal(t, "./data/course_creator.db", config.Path)
	assert.False(t, config.Debug)
	assert.Equal(t, "development", config.Env)
}

func TestNewDatabase_WithNilConfig(t *testing.T) {
	// Use in-memory database for testing
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(&Config{
		Path:  dbPath,
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test Ping
	err = db.Ping()
	assert.NoError(t, err)

	// Test GetGormDB
	gormDB := db.GetGormDB()
	assert.NotNil(t, gormDB)

	// Test Close
	err = db.Close()
	assert.NoError(t, err)
}

func TestNewDatabase_WithValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := &Config{
		Path:  dbPath,
		Debug: true,
		Env:   "test",
	}

	db, err := NewDatabase(config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify database file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)

	// Test Ping
	err = db.Ping()
	assert.NoError(t, err)

	// Test GetGormDB
	gormDB := db.GetGormDB()
	assert.NotNil(t, gormDB)
}

func TestNewDatabase_DataDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	dbPath := filepath.Join(dataDir, "course_creator.db")

	// Ensure data directory doesn't exist initially
	_, err := os.Stat(dataDir)
	assert.True(t, os.IsNotExist(err))

	config := &Config{
		Path:  dbPath,
		Debug: false,
		Env:   "test",
	}

	db, err := NewDatabase(config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify data directory was created
	_, err = os.Stat(dataDir)
	assert.NoError(t, err)
	assert.True(t, func() bool { stat, _ := os.Stat(dataDir); return stat.IsDir() }())
}

func TestNewDatabase_InvalidPath(t *testing.T) {
	// Try to create database in a directory that can't be created
	config := &Config{
		Path:  "/invalid/path/that/cannot/be/created/database.db",
		Debug: false,
		Env:   "test",
	}

	db, err := NewDatabase(config)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to create data directory")
}

func TestDB_Close(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(&Config{
		Path:  dbPath,
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	// Close should work
	err = db.Close()
	assert.NoError(t, err)

	// Ping after close should fail
	err = db.Ping()
	assert.Error(t, err)
}

func TestDB_Ping(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(&Config{
		Path:  dbPath,
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Ping should succeed
	err = db.Ping()
	assert.NoError(t, err)
}

func TestDB_GetGormDB(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDatabase(&Config{
		Path:  dbPath,
		Debug: false,
		Env:   "test",
	})
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// GetGormDB should return a valid GORM instance
	gormDB := db.GetGormDB()
	assert.NotNil(t, gormDB)

	// Should be able to perform operations on it
	err = gormDB.AutoMigrate() // This should not error even if no models
	assert.NoError(t, err)
}
