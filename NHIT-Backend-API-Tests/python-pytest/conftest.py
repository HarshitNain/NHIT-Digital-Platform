"""
Pytest configuration and fixtures
"""
import pytest
import os
from dotenv import load_dotenv
from utils.api_client import APIClient
from utils.test_data import TestDataFactory

# Load environment variables
load_dotenv()


@pytest.fixture(scope="session")
def api_base_url():
    """Base URL for API"""
    return os.getenv("API_BASE_URL", "http://localhost:8080")


@pytest.fixture(scope="session")
def api_client(api_base_url):
    """API client instance"""
    return APIClient(base_url=api_base_url)


@pytest.fixture(scope="function")
def test_data():
    """Test data factory"""
    return TestDataFactory()


@pytest.fixture(scope="function")
def test_user(api_client, test_data):
    """Create a test user and clean up after test"""
    user_data = test_data.create_user_data()
    
    # Create user
    response = api_client.post("/api/v1/users", json=user_data)
    user = response.json()
    
    yield user
    
    # Cleanup: Delete user after test
    try:
        api_client.delete(f"/api/v1/users/{user['user_id']}")
    except:
        pass  # Ignore cleanup errors


@pytest.fixture(scope="function")
def authenticated_client(api_client, test_user):
    """API client with authentication token"""
    # Login
    login_data = {
        "login": test_user["email"],
        "password": "testpass123",
        "tenant_id": test_user.get("tenant_id", os.getenv("TEST_TENANT_ID"))
    }
    
    response = api_client.post("/api/v1/auth/login", json=login_data)
    token = response.json().get("token")
    
    # Set auth token
    api_client.set_auth_token(token)
    
    yield api_client
    
    # Logout
    try:
        api_client.post("/api/v1/auth/logout", json={"user_id": test_user["user_id"]})
    except:
        pass


@pytest.fixture(autouse=True)
def reset_client_state(api_client):
    """Reset client state before each test"""
    api_client.clear_auth_token()
    yield
    api_client.clear_auth_token()


def pytest_configure(config):
    """Pytest configuration hook"""
    # Create reports directory
    os.makedirs("reports", exist_ok=True)
    
    # Register custom markers
    config.addinivalue_line("markers", "smoke: Quick smoke tests")
    config.addinivalue_line("markers", "regression: Full regression suite")
    config.addinivalue_line("markers", "integration: Integration tests")


def pytest_collection_modifyitems(config, items):
    """Modify test collection"""
    # Add markers based on test file names
    for item in items:
        if "test_auth" in item.nodeid:
            item.add_marker(pytest.mark.auth)
        elif "test_users" in item.nodeid:
            item.add_marker(pytest.mark.users)
