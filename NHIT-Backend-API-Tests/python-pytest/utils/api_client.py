"""
API Client for making HTTP requests
"""
import requests
import logging
from typing import Dict, Any, Optional

logger = logging.getLogger(__name__)


class APIClient:
    """HTTP client for API testing"""
    
    def __init__(self, base_url: str, timeout: int = 30):
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })
    
    def set_auth_token(self, token: str):
        """Set authentication token"""
        self.session.headers['Authorization'] = f'Bearer {token}'
        logger.info("Auth token set")
    
    def clear_auth_token(self):
        """Clear authentication token"""
        self.session.headers.pop('Authorization', None)
        logger.info("Auth token cleared")
    
    def get(self, endpoint: str, **kwargs) -> requests.Response:
        """GET request"""
        url = f"{self.base_url}{endpoint}"
        logger.info(f"GET {url}")
        response = self.session.get(url, timeout=self.timeout, **kwargs)
        logger.info(f"Response: {response.status_code}")
        return response
    
    def post(self, endpoint: str, **kwargs) -> requests.Response:
        """POST request"""
        url = f"{self.base_url}{endpoint}"
        logger.info(f"POST {url}")
        response = self.session.post(url, timeout=self.timeout, **kwargs)
        logger.info(f"Response: {response.status_code}")
        return response
    
    def put(self, endpoint: str, **kwargs) -> requests.Response:
        """PUT request"""
        url = f"{self.base_url}{endpoint}"
        logger.info(f"PUT {url}")
        response = self.session.put(url, timeout=self.timeout, **kwargs)
        logger.info(f"Response: {response.status_code}")
        return response
    
    def delete(self, endpoint: str, **kwargs) -> requests.Response:
        """DELETE request"""
        url = f"{self.base_url}{endpoint}"
        logger.info(f"DELETE {url}")
        response = self.session.delete(url, timeout=self.timeout, **kwargs)
        logger.info(f"Response: {response.status_code}")
        return response
    
    def patch(self, endpoint: str, **kwargs) -> requests.Response:
        """PATCH request"""
        url = f"{self.base_url}{endpoint}"
        logger.info(f"PATCH {url}")
        response = self.session.patch(url, timeout=self.timeout, **kwargs)
        logger.info(f"Response: {response.status_code}")
        return response
