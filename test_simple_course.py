import requests
import json
import time

# API base URL
BASE_URL = "http://localhost:8080/api/v1"

def test_course_generation():
    """Test course generation directly without checking job status"""
    
    # Load saved tokens
    with open("test_tokens.json", "r") as f:
        tokens = json.load(f)
    access_token = tokens["access_token"]
    
    # 1. Generate a course
    print("Generating course...")
    course_response = requests.post(
        f"{BASE_URL}/courses/generate",
        headers={"Authorization": f"Bearer {access_token}"},
        json={
            "markdown": """# My Test Course

## Lesson 1: Introduction
This is the introduction to the course.

## Lesson 2: Getting Started  
Here's how to get started.

## Lesson 3: Advanced Topics
Let's explore advanced concepts.
""",
            "options": {
                "quality": "high",
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
    
    # 2. Wait a bit for processing
    print("\nWaiting 10 seconds for course generation to complete...")
    time.sleep(10)
    
    # 3. Check if course was created
    print("\nListing courses to see if new course appears...")
    courses_response = requests.get(
        f"{BASE_URL}/courses",
        headers={"Authorization": f"Bearer {access_token}"}
    )
    
    if courses_response.status_code == 200:
        courses = courses_response.json()
        print(f"Total courses: {len(courses)}")
        for course in courses:
            print(f"  - {course.get('title')} (ID: {course.get('id')})")
            
            # Get lessons for this course
            course_detail = requests.get(
                f"{BASE_URL}/courses/{course.get('id')}",
                headers={"Authorization": f"Bearer {access_token}"}
            )
            
            if course_detail.status_code == 200:
                course = course_detail.json()
                lessons = course.get('lessons', [])
                print(f"    Lessons: {len(lessons)}")
                for lesson in lessons:
                    print(f"      - {lesson.get('title')}")

if __name__ == "__main__":
    test_course_generation()