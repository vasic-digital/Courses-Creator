-- Remove course_metadata table
-- Migration: 002_add_course_metadata.down.sql

DROP TRIGGER IF EXISTS update_course_metadata_updated_at ON course_metadata;
DROP TABLE IF EXISTS course_metadata;