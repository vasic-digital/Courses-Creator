import requests

def test_course_api():
    base_url = "http://localhost:8080/api/v1"
    
    # First login to get token
    login_data = {
        "email": "test@example.com",
        "password": "password123"
    }
    
    response = requests.post(f"{base_url}/auth/login", json=login_data)
    if response.status_code != 200:
        print("Login failed, trying to register first...")
        # Register user
        register_data = {
            "email": "test@example.com",
            "password": "password123",
            "first_name": "Test",
            "last_name": "User"
        }
        response = requests.post(f"{base_url}/auth/register", json=register_data)
        print(f"Register response: {response.status_code}")
        if response.status_code != 201:
            print(f"Registration failed: {response.json()}")
            return
            
        # Try login again
        response = requests.post(f"{base_url}/auth/login", json=login_data)
    
    # Extract access token
    access_token = response.json().get("access_token", "")
    headers = {"Authorization": f"Bearer {access_token}"}
    
    # Test course generation
    print("\nTesting course generation...")
    course_data = {
        "markdown": "# Test Course\n\nThis is a test course.\n\n## Lesson 1\n\nContent for lesson 1.",
        "options": {
            "quality": "standard",
            "languages": ["en"]
        }
    }
    
    response = requests.post(f"{base_url}/courses/generate", json=course_data, headers=headers)
    print(f"Generate course response status: {response.status_code}")
    print(f"Generate course response: {response.json()}")
    
    # Test courses list
    print("\nTesting courses list...")
    response = requests.get(f"{base_url}/courses", headers=headers)
    print(f"Courses list response status: {response.status_code}")
    print(f"Courses list response: {response.json()}")
    
    # Test course details
    if response.status_code == 200:
        # Handle both list and object responses
        course_data = response.json()
        if isinstance(course_data, dict) and "courses" in course_data:
            courses = course_data["courses"]
        elif isinstance(course_data, list):
            courses = course_data
        else:
            courses = []
            
        if courses:
            course_id = courses[0].get("id")
            if course_id:
                print(f"\nTesting course details for ID: {course_id}")
                response = requests.get(f"{base_url}/courses/{course_id}", headers=headers)
                print(f"Course details response status: {response.status_code}")
                print(f"Course details response: {response.json()}")

if __name__ == "__main__":
    test_course_api()