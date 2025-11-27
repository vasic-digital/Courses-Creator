import requests
import json
import time

# API base URL
BASE_URL = "http://localhost:8080/api/v1"

def test_course_generation_no_audio():
    """Test course generation without audio"""
    
    # Load saved tokens
    with open("test_tokens.json", "r") as f:
        tokens = json.load(f)
    access_token = tokens["access_token"]
    
    # 1. Generate a course with quality set to skip TTS if possible
    print("Generating course (trying to skip audio)...")
    course_response = requests.post(
        f"{BASE_URL}/courses/generate",
        headers={"Authorization": f"Bearer {access_token}"},
        json={
            "markdown": """# Simple Test Course

## Lesson 1: Basic Introduction
This is a simple test lesson.

## Lesson 2: Another Lesson
This is another test lesson.
""",
            "options": {
                "quality": "draft",  # Try using draft quality
                "background_music": False,
                "languages": ["en"]
            }
        }
    )
    
    if course_response.status_code != 202:
        print(f"Course generation failed: {course_response.text}")
        return
    
    course_data = course_response.json()
    job_id = course_data["job_id"]
    print(f"Course generation started with job ID: {job_id}")
    
    # 2. Check processing job directly in database
    print("\nChecking processing jobs in database...")
    import subprocess
    try:
        # Query the processing_jobs table
        result = subprocess.run(
            ["sqlite3", "/Volumes/T7/Projects/Course-Creator/core-processor/data/course_creator.db",
             f"SELECT id, status, progress, error FROM processing_jobs WHERE id = '{job_id}';"],
            capture_output=True, text=True
        )
        if result.stdout:
            print("Processing job:")
            print(result.stdout)
    except Exception as e:
        print(f"Failed to query database: {e}")
    
    # 3. Check if we have a processing job endpoint
    print("\nTrying to check processing job status via API...")
    # First, let's see what endpoints are available for jobs
    response = requests.get(
        f"{BASE_URL}/jobs",
        headers={"Authorization": f"Bearer {access_token}"}
    )
    
    if response.status_code == 200:
        jobs = response.json()
        print(f"User jobs: {jobs}")
    
    # 4. Wait a bit and check again
    print("\nWaiting 5 seconds...")
    time.sleep(5)
    
    # 5. Query database again
    try:
        result = subprocess.run(
            ["sqlite3", "/Volumes/T7/Projects/Course-Creator/core-processor/data/course_creator.db",
             f"SELECT id, status, progress, error FROM processing_jobs WHERE id = '{job_id}';"],
            capture_output=True, text=True
        )
        if result.stdout:
            print("Processing job after 5 seconds:")
            print(result.stdout)
    except Exception as e:
        print(f"Failed to query database: {e}")

if __name__ == "__main__":
    test_course_generation_no_audio()