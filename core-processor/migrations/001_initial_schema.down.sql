-- Rollback for initial schema
-- Migration: 001_initial_schema.down.sql

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS lessons;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS users;

-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();