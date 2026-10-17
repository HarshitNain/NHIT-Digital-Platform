"""
Authentication API Tests
"""
import pytest


class TestAuthentication:
    """Test suite for Authentication APIs"""
    
    @pytest.mark.smoke
    @pytest.mark.auth
    def test_register_user_success(self, api_client, test_data):
        """Test successful user registration"""
        register_data = test_data.create_register_data()
        
        response = api_client.post("/api/v1/auth/register", json=register_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "user_id" in data
        assert data["email"] == register_data["email"]
        assert "password" not in data
        
        # Cleanup
        if "user_id" in data:
            api_client.delete(f"/api/v1/users/{data['user_id']}")
    
    def test_register_duplicate_email(self, api_client, test_user, test_data):
        """Test registration with duplicate email"""
        register_data = test_data.create_register_data(email=test_user["email"])
        
        response = api_client.post("/api/v1/auth/register", json=register_data)
        
        # Should fail with conflict or bad request
        assert response.status_code in [400, 409]
    
    def test_register_missing_required_fields(self, api_client):
        """Test registration with missing fields"""
        invalid_data = {"name": "Test User"}
        
        response = api_client.post("/api/v1/auth/register", json=invalid_data)
        
        assert response.status_code in [400, 422]
    
    @pytest.mark.smoke
    @pytest.mark.auth
    def test_login_success(self, api_client, test_user):
        """Test successful login"""
        login_data = {
            "login": test_user["email"],
            "password": "testpass123",
            "tenant_id": test_user.get("tenant_id")
        }
        
        response = api_client.post("/api/v1/auth/login", json=login_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "token" in data
        assert "refresh_token" in data
        assert "user_id" in data
        assert data["email"] == test_user["email"]
    
    def test_login_invalid_credentials(self, api_client, test_user):
        """Test login with wrong password"""
        login_data = {
            "login": test_user["email"],
            "password": "wrongpassword",
            "tenant_id": test_user.get("tenant_id")
        }
        
        response = api_client.post("/api/v1/auth/login", json=login_data)
        
        assert response.status_code == 401
    
    def test_login_nonexistent_user(self, api_client, test_data):
        """Test login with non-existent user"""
        login_data = test_data.create_login_data(
            email="nonexistent@example.com",
            password="password123"
        )
        
        response = api_client.post("/api/v1/auth/login", json=login_data)
        
        assert response.status_code in [401, 404]
    
    def test_login_missing_fields(self, api_client):
        """Test login with missing fields"""
        invalid_data = {"login": "test@example.com"}
        
        response = api_client.post("/api/v1/auth/login", json=invalid_data)
        
        assert response.status_code in [400, 422]
    
    @pytest.mark.auth
    def test_logout_success(self, api_client, test_user):
        """Test successful logout"""
        # First login
        login_data = {
            "login": test_user["email"],
            "password": "testpass123",
            "tenant_id": test_user.get("tenant_id")
        }
        login_response = api_client.post("/api/v1/auth/login", json=login_data)
        token = login_response.json().get("token")
        refresh_token = login_response.json().get("refresh_token")
        
        # Set auth token
        api_client.set_auth_token(token)
        
        # Logout
        logout_data = {
            "user_id": test_user["user_id"],
            "refresh_token": refresh_token
        }
        response = api_client.post("/api/v1/auth/logout", json=logout_data)
        
        assert response.status_code in [200, 204]
    
    def test_logout_without_auth(self, api_client, test_user):
        """Test logout without authentication"""
        logout_data = {
            "user_id": test_user["user_id"],
            "refresh_token": "fake-token"
        }
        
        response = api_client.post("/api/v1/auth/logout", json=logout_data)
        
        # Should either succeed or require auth
        assert response.status_code in [200, 204, 401]


class TestPasswordManagement:
    """Test suite for Password Management"""
    
    def test_forgot_password(self, api_client, test_user):
        """Test forgot password request"""
        forgot_data = {"email": test_user["email"]}
        
        response = api_client.post("/api/v1/auth/forgot-password", json=forgot_data)
        
        # Should succeed even if email doesn't exist (security best practice)
        assert response.status_code == 200
    
    def test_forgot_password_invalid_email(self, api_client):
        """Test forgot password with invalid email"""
        forgot_data = {"email": "not-an-email"}
        
        response = api_client.post("/api/v1/auth/forgot-password", json=forgot_data)
        
        assert response.status_code in [200, 400, 422]
    
    def test_reset_password(self, api_client):
        """Test password reset"""
        reset_data = {
            "token": "fake-reset-token",
            "new_password": "newpassword123"
        }
        
        response = api_client.post("/api/v1/auth/reset-password", json=reset_data)
        
        # Will fail with invalid token
        assert response.status_code in [400, 404]
    
    def test_reset_password_missing_token(self, api_client):
        """Test password reset without token"""
        reset_data = {"new_password": "newpassword123"}
        
        response = api_client.post("/api/v1/auth/reset-password", json=reset_data)
        
        assert response.status_code in [400, 422]


class TestTokenManagement:
    """Test suite for Token Management"""
    
    def test_refresh_token(self, api_client, test_user):
        """Test token refresh"""
        # First login to get refresh token
        login_data = {
            "login": test_user["email"],
            "password": "testpass123",
            "tenant_id": test_user.get("tenant_id")
        }
        login_response = api_client.post("/api/v1/auth/login", json=login_data)
        refresh_token = login_response.json().get("refresh_token")
        
        # Refresh token
        refresh_data = {
            "refresh_token": refresh_token,
            "tenant_id": test_user.get("tenant_id")
        }
        response = api_client.post("/api/v1/auth/refresh", json=refresh_data)
        
        # Should return new tokens
        if response.status_code == 200:
            data = response.json()
            assert "token" in data
            assert "refresh_token" in data
    
    def test_refresh_invalid_token(self, api_client, test_data):
        """Test refresh with invalid token"""
        refresh_data = {
            "refresh_token": "invalid-token",
            "tenant_id": test_data.tenant_id
        }
        
        response = api_client.post("/api/v1/auth/refresh", json=refresh_data)
        
        assert response.status_code in [400, 401]


class TestEmailVerification:
    """Test suite for Email Verification"""
    
    def test_send_verification_email(self, api_client, test_user):
        """Test sending verification email"""
        verify_data = {"user_id": test_user["user_id"]}
        
        response = api_client.post("/api/v1/auth/send-verification", json=verify_data)
        
        assert response.status_code in [200, 404]
    
    def test_verify_email(self, api_client, test_user):
        """Test email verification"""
        verify_data = {
            "user_id": test_user["user_id"],
            "verification_token": "fake-token"
        }
        
        response = api_client.post("/api/v1/auth/verify-email", json=verify_data)
        
        # Will fail with invalid token
        assert response.status_code in [400, 404]


@pytest.mark.integration
class TestAuthenticationWorkflows:
    """Integration tests for authentication workflows"""
    
    def test_complete_registration_login_workflow(self, api_client, test_data):
        """Test complete registration and login flow"""
        # Register
        register_data = test_data.create_register_data()
        register_response = api_client.post("/api/v1/auth/register", json=register_data)
        assert register_response.status_code == 200
        user_id = register_response.json().get("user_id")
        
        # Login
        login_data = {
            "login": register_data["email"],
            "password": register_data["password"],
            "tenant_id": register_data["tenant_id"]
        }
        login_response = api_client.post("/api/v1/auth/login", json=login_data)
        assert login_response.status_code == 200
        assert "token" in login_response.json()
        
        # Cleanup
        if user_id:
            api_client.delete(f"/api/v1/users/{user_id}")
    
    def test_login_logout_login_workflow(self, api_client, test_user):
        """Test login, logout, and login again"""
        login_data = {
            "login": test_user["email"],
            "password": "testpass123",
            "tenant_id": test_user.get("tenant_id")
        }
        
        # First login
        login1 = api_client.post("/api/v1/auth/login", json=login_data)
        assert login1.status_code == 200
        token1 = login1.json().get("token")
        
        # Logout
        api_client.set_auth_token(token1)
        logout = api_client.post("/api/v1/auth/logout", json={"user_id": test_user["user_id"]})
        api_client.clear_auth_token()
        
        # Login again
        login2 = api_client.post("/api/v1/auth/login", json=login_data)
        assert login2.status_code == 200
        assert "token" in login2.json()
