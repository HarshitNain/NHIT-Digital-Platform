"""
Test data factory for generating test data
"""
import os
from faker import Faker
from typing import Dict, Any

fake = Faker()


class TestDataFactory:
    """Factory for creating test data"""
    
    def __init__(self):
        self.tenant_id = os.getenv("TEST_TENANT_ID", "123e4567-e89b-12d3-a456-426614174000")
    
    def create_user_data(self, **overrides) -> Dict[str, Any]:
        """Generate user data"""
        data = {
            "tenant_id": self.tenant_id,
            "name": fake.name(),
            "email": fake.email(),
            "password": "testpass123"
        }
        data.update(overrides)
        return data
    
    def create_login_data(self, email: str = None, password: str = None) -> Dict[str, Any]:
        """Generate login data"""
        return {
            "login": email or fake.email(),
            "password": password or "testpass123",
            "tenant_id": self.tenant_id
        }
    
    def create_register_data(self, **overrides) -> Dict[str, Any]:
        """Generate registration data"""
        data = {
            "tenant_id": self.tenant_id,
            "name": fake.name(),
            "email": fake.email(),
            "password": "testpass123",
            "roles": ["USER"]
        }
        data.update(overrides)
        return data
    
    def create_update_user_data(self, **overrides) -> Dict[str, Any]:
        """Generate user update data"""
        data = {
            "name": fake.name(),
            "email": fake.email()
        }
        data.update(overrides)
        return data
    
    def create_role_assignment_data(self, role_ids: list = None) -> Dict[str, Any]:
        """Generate role assignment data"""
        return {
            "roles": role_ids or ["role-uuid-1", "role-uuid-2"]
        }
