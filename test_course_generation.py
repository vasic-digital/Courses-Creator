import requests
import json
import time
import os

# API base URL
BASE_URL = "http://localhost:8080/api/v1"

def test_complete_flow():
    """Test the complete course generation flow"""
    
    # Try to load saved tokens
    if os.path.exists("test_tokens.json"):
        with open("test_tokens.json", "r") as f:
            tokens = json.load(f)
        access_token = tokens["access_token"]
        print("Using saved access token")
    else:
        # 1. Login to get tokens
        print("Logging in...")
        login_response = requests.post(
            f"{BASE_URL}/auth/login",
            json={
                "email": "testuser@example.com",
                "password": "testpassword123"
            }
        )
        
        if login_response.status_code != 200:
            print(f"Login failed: {login_response.text}")
            return
        
        tokens = login_response.json()
        access_token = tokens["access_token"]
    
    # 2. Generate a course
    print("\nGenerating course...")
    course_response = requests.post(
        f"{BASE_URL}/courses/generate",
        headers={"Authorization": f"Bearer {access_token}"},
        json={
            "markdown": """# My Test Course

## Lesson 1: Introduction
This is the introduction to the course.

## Lesson 2: Getting Started
Here's how to get started with the topic.

## Lesson 3: Advanced Topics
Let's explore some advanced concepts.
""",
            "options": {
                "voice": "v2/en_speaker_6",
                "background_music": False,
                "quality": "high"
            }
        }
    )
    
    if course_response.status_code != 202:
        print(f"Course generation failed: {course_response.text}")
        return
    
    course_data = course_response.json()
    job_id = course_data["job_id"]
    print(f"Course generation started with job ID: {job_id}")
    
    # 3. Monitor job progress
    print("\nMonitoring job progress...")
    while True:
        job_response = requests.get(
            f"{BASE_URL}/jobs/{job_id}",
            headers={"Authorization": f"Bearer {access_token}"}
        )
        
        if job_response.status_code != 200:
            print(f"Failed to get job status: {job_response.text}")
            break
        
        job_data = job_response.json()
        status = job_data.get("status", "unknown")
        progress = job_data.get("progress", 0)
        error = job_data.get("error", None)
        
        print(f"Job status: {status}, Progress: {progress}%")
        
        if status == "completed":
            print("\nCourse generation completed!")
            course_id = job_data.get("course_id")
            if course_id:
                print(f"Generated course ID: {course_id}")
                
                # 4. Get course details
                course_detail = requests.get(
                    f"{BASE_URL}/courses/{course_id}",
                    headers={"Authorization": f"Bearer {access_token}"}
                )
                
                if course_detail.status_code == 200:
                    course = course_detail.json()
                    print(f"\nCourse Details:")
                    print(f"  Title: {course.get('title')}")
                    print(f"  Description: {course.get('description')}")
                    print(f"  Lessons: {len(course.get('lessons', []))}")
            break
        elif status == "failed" or error:
            print(f"\nJob failed: {error}")
            break
        
        time.sleep(2)  # Poll every 2 seconds
    
    # 5. List all courses
    print("\nListing all courses...")
    courses_response = requests.get(
        f"{BASE_URL}/courses",
        headers={"Authorization": f"Bearer {access_token}"}
    )
    
    if courses_response.status_code == 200:
        courses = courses_response.json()
        print(f"Total courses: {len(courses)}")
        for course in courses:
            print(f"  - {course.get('title')} (ID: {course.get('id')})")

if __name__ == "__main__":
    test_complete_flow()