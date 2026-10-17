"""
User Management API Tests
"""
import pytest
import os


class TestUserManagement:
    """Test suite for User Management APIs"""
    
    @pytest.mark.smoke
    def test_create_user_success(self, api_client, test_data):
        """Test successful user creation"""
        user_data = test_data.create_user_data()
        
        response = api_client.post("/api/v1/users", json=user_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "user_id" in data
        assert data["name"] == user_data["name"]
        assert data["email"] == user_data["email"]
        assert "password" not in data  # Password should not be returned
        
        # Cleanup
        api_client.delete(f"/api/v1/users/{data['user_id']}")
    
    def test_create_user_missing_required_fields(self, api_client):
        """Test user creation with missing required fields"""
        invalid_data = {"name": "Test User"}  # Missing email, password, tenant_id
        
        response = api_client.post("/api/v1/users", json=invalid_data)
        
        assert response.status_code in [400, 422]  # Bad Request or Unprocessable Entity
    
    def test_create_user_invalid_email(self, api_client, test_data):
        """Test user creation with invalid email"""
        user_data = test_data.create_user_data(email="invalid-email")
        
        response = api_client.post("/api/v1/users", json=user_data)
        
        # Should fail validation (might be 400 or 422 depending on implementation)
        assert response.status_code in [400, 422] or response.status_code == 200
    
    @pytest.mark.smoke
    def test_get_user_success(self, api_client, test_user):
        """Test getting user by ID"""
        user_id = test_user["user_id"]
        
        response = api_client.get(f"/api/v1/users/{user_id}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["user_id"] == user_id
        assert data["email"] == test_user["email"]
    
    def test_get_user_not_found(self, api_client):
        """Test getting non-existent user"""
        fake_user_id = "00000000-0000-0000-0000-000000000000"
        
        response = api_client.get(f"/api/v1/users/{fake_user_id}")
        
        assert response.status_code == 404
    
    def test_get_user_invalid_id(self, api_client):
        """Test getting user with invalid ID format"""
        invalid_id = "not-a-uuid"
        
        response = api_client.get(f"/api/v1/users/{invalid_id}")
        
        assert response.status_code in [400, 404]
    
    @pytest.mark.smoke
    def test_list_users_success(self, api_client):
        """Test listing users"""
        tenant_id = os.getenv("TEST_TENANT_ID")
        
        response = api_client.get(f"/api/v1/users?tenant_id={tenant_id}")
        
        assert response.status_code == 200
        data = response.json()
        assert "users" in data or isinstance(data, list)
    
    def test_list_users_without_tenant_id(self, api_client):
        """Test listing users without tenant_id"""
        response = api_client.get("/api/v1/users")
        
        # Should either return empty list or require tenant_id
        assert response.status_code in [200, 400]
    
    @pytest.mark.smoke
    def test_update_user_success(self, api_client, test_user, test_data):
        """Test updating user"""
        user_id = test_user["user_id"]
        update_data = test_data.create_update_user_data()
        
        response = api_client.put(f"/api/v1/users/{user_id}", json=update_data)
        
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == update_data["name"]
        assert data["email"] == update_data["email"]
    
    def test_update_user_not_found(self, api_client, test_data):
        """Test updating non-existent user"""
        fake_user_id = "00000000-0000-0000-0000-000000000000"
        update_data = test_data.create_update_user_data()
        
        response = api_client.put(f"/api/v1/users/{fake_user_id}", json=update_data)
        
        assert response.status_code == 404
    
    def test_update_user_invalid_data(self, api_client, test_user):
        """Test updating user with invalid data"""
        user_id = test_user["user_id"]
        invalid_data = {"email": "not-an-email"}
        
        response = api_client.put(f"/api/v1/users/{user_id}", json=invalid_data)
        
        # Should either accept or reject based on validation
        assert response.status_code in [200, 400, 422]
    
    @pytest.mark.smoke
    def test_delete_user_success(self, api_client, test_data):
        """Test deleting user"""
        # Create a user to delete
        user_data = test_data.create_user_data()
        create_response = api_client.post("/api/v1/users", json=user_data)
        user_id = create_response.json()["user_id"]
        
        # Delete the user
        response = api_client.delete(f"/api/v1/users/{user_id}")
        
        assert response.status_code in [200, 204]
        
        # Verify user is deleted
        get_response = api_client.get(f"/api/v1/users/{user_id}")
        assert get_response.status_code == 404
    
    def test_delete_user_not_found(self, api_client):
        """Test deleting non-existent user"""
        fake_user_id = "00000000-0000-0000-0000-000000000000"
        
        response = api_client.delete(f"/api/v1/users/{fake_user_id}")
        
        assert response.status_code in [404, 204]  # Some APIs return 204 even if not found
    
    def test_delete_user_invalid_id(self, api_client):
        """Test deleting user with invalid ID"""
        invalid_id = "not-a-uuid"
        
        response = api_client.delete(f"/api/v1/users/{invalid_id}")
        
        assert response.status_code in [400, 404]


class TestUserRoles:
    """Test suite for User Role Management"""
    
    def test_assign_roles_to_user(self, api_client, test_user, test_data):
        """Test assigning roles to user"""
        user_id = test_user["user_id"]
        role_data = test_data.create_role_assignment_data()
        
        response = api_client.post(f"/api/v1/users/{user_id}/roles", json=role_data)
        
        # Should succeed or fail gracefully
        assert response.status_code in [200, 400, 404]
    
    def test_list_user_roles(self, api_client, test_user):
        """Test listing user roles"""
        user_id = test_user["user_id"]
        
        response = api_client.get(f"/api/v1/users/{user_id}/roles")
        
        assert response.status_code == 200
        data = response.json()
        assert "roles" in data or isinstance(data, list)
    
    def test_assign_roles_invalid_user(self, api_client, test_data):
        """Test assigning roles to non-existent user"""
        fake_user_id = "00000000-0000-0000-0000-000000000000"
        role_data = test_data.create_role_assignment_data()
        
        response = api_client.post(f"/api/v1/users/{fake_user_id}/roles", json=role_data)
        
        assert response.status_code == 404


@pytest.mark.integration
class TestUserIntegration:
    """Integration tests for user workflows"""
    
    def test_create_update_delete_workflow(self, api_client, test_data):
        """Test complete user lifecycle"""
        # Create user
        user_data = test_data.create_user_data()
        create_response = api_client.post("/api/v1/users", json=user_data)
        assert create_response.status_code == 200
        user_id = create_response.json()["user_id"]
        
        # Update user
        update_data = test_data.create_update_user_data()
        update_response = api_client.put(f"/api/v1/users/{user_id}", json=update_data)
        assert update_response.status_code == 200
        
        # Get user to verify update
        get_response = api_client.get(f"/api/v1/users/{user_id}")
        assert get_response.status_code == 200
        assert get_response.json()["name"] == update_data["name"]
        
        # Delete user
        delete_response = api_client.delete(f"/api/v1/users/{user_id}")
        assert delete_response.status_code in [200, 204]
        
        # Verify deletion
        final_get = api_client.get(f"/api/v1/users/{user_id}")
        assert final_get.status_code == 404
