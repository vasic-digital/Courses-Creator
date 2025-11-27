import requests

def test_auth():
    base_url = "http://localhost:8080/api/v1"
    
    # Test registration
    print("Testing registration...")
    register_data = {
        "email": "test@example.com",
        "password": "password123",
        "first_name": "Test",
        "last_name": "User"
    }
    
    response = requests.post(f"{base_url}/auth/register", json=register_data)
    print(f"Register response status: {response.status_code}")
    print(f"Register response: {response.json()}")
    
    # Test login
    print("\nTesting login...")
    login_data = {
        "email": "test@example.com",
        "password": "password123"
    }
    
    response = requests.post(f"{base_url}/auth/login", json=login_data)
    print(f"Login response status: {response.status_code}")
    print(f"Login response: {response.json()}")
    
    # Extract access token
    access_token = response.json().get("access_token", "")
    
    if access_token:
        print("\nTesting protected endpoint with valid token...")
        headers = {"Authorization": f"Bearer {access_token}"}
        
        # Test user profile
        response = requests.get(f"{base_url}/auth/profile", headers=headers)
        print(f"Profile response status: {response.status_code}")
        print(f"Profile response: {response.json()}")
        
        # Test courses list
        response = requests.get(f"{base_url}/courses", headers=headers)
        print(f"Courses list response status: {response.status_code}")
        print(f"Courses list response: {response.json()}")
        
        print("\nTesting protected endpoint without token...")
        response = requests.get(f"{base_url}/auth/profile")
        print(f"Profile without token status: {response.status_code}")
        print(f"Profile without token response: {response.json()}")

if __name__ == "__main__":
    test_auth()