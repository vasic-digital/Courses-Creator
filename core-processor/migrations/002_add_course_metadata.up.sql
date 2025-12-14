-- Add course_metadata table
-- Migration: 002_add_course_metadata.up.sql

-- Course metadata table for storing additional course properties
CREATE TABLE IF NOT EXISTS course_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    author VARCHAR(255),
    language VARCHAR(10),
    tags TEXT, -- JSON string
    thumbnail_url VARCHAR(500),
    total_duration INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_course_metadata_course_id ON course_metadata(course_id);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_course_metadata_updated_at BEFORE UPDATE ON course_metadata
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();