import requests
import json

# API base URL
BASE_URL = "http://localhost:8080/api/v1"

def register_user():
    """Register a new user"""
    print("Registering user...")
    response = requests.post(
        f"{BASE_URL}/auth/register",
        json={
            "email": "testuser@example.com",
            "password": "testpassword123",
            "first_name": "Test",
            "last_name": "User"
        }
    )
    
    if response.status_code == 200:
        print("User registered successfully")
        return True
    elif response.status_code == 409:
        print("User already exists")
        return True
    else:
        print(f"Registration failed: {response.text}")
        return False

def login_user():
    """Login and get tokens"""
    print("\nLogging in...")
    response = requests.post(
        f"{BASE_URL}/auth/login",
        json={
            "email": "testuser@example.com",
            "password": "testpassword123"
        }
    )
    
    if response.status_code == 200:
        tokens = response.json()
        print("Login successful")
        return tokens["access_token"]
    else:
        print(f"Login failed: {response.text}")
        return None

if __name__ == "__main__":
    if register_user():
        token = login_user()
        if token:
            print(f"\nAccess token: {token[:50]}...")
            # Save token to file for later use
            with open("access_token.txt", "w") as f:
                f.write(token)
            print("Token saved to access_token.txt")