# Python pytest API Tests

## 🎯 Overview

Comprehensive API testing suite using pytest for NHIT Backend REST APIs.

## 📦 Installation

```bash
# Create virtual environment
python -m venv venv

# Activate virtual environment
# Windows
venv\Scripts\activate
# Linux/Mac
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt
```

## ⚙️ Configuration

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Update `.env` with your configuration:
```env
API_BASE_URL=http://localhost:8080
TEST_TENANT_ID=your-tenant-id
TEST_USER_EMAIL=test@example.com
TEST_USER_PASSWORD=testpass123
```

## 🚀 Running Tests

### Run All Tests
```bash
pytest
```

### Run Specific Test File
```bash
pytest tests/test_users.py
pytest tests/test_auth.py
```

### Run Specific Test Class
```bash
pytest tests/test_users.py::TestUserManagement
```

### Run Specific Test
```bash
pytest tests/test_users.py::TestUserManagement::test_create_user_success
```

### Run Tests by Marker
```bash
# Run only smoke tests
pytest -m smoke

# Run only auth tests
pytest -m auth

# Run integration tests
pytest -m integration
```

### Run Tests in Parallel
```bash
pytest -n 4  # Run with 4 workers
```

### Generate HTML Report
```bash
pytest --html=reports/report.html --self-contained-html
```

### Run with Coverage
```bash
pytest --cov=. --cov-report=html
```

### Verbose Output
```bash
pytest -v
```

### Show Print Statements
```bash
pytest -s
```

## 📊 Test Structure

```
tests/
├── test_users.py           # User management tests
├── test_auth.py            # Authentication tests
└── test_integration.py     # Integration tests

utils/
├── api_client.py           # HTTP client wrapper
└── test_data.py            # Test data factory
```

## 🏷️ Test Markers

- `@pytest.mark.smoke` - Quick smoke tests
- `@pytest.mark.regression` - Full regression suite
- `@pytest.mark.integration` - Integration tests
- `@pytest.mark.auth` - Authentication tests
- `@pytest.mark.users` - User management tests
- `@pytest.mark.slow` - Slow running tests

## 📝 Writing Tests

### Basic Test Example
```python
def test_create_user(api_client, test_data):
    """Test user creation"""
    user_data = test_data.create_user_data()
    response = api_client.post("/api/v1/users", json=user_data)
    
    assert response.status_code == 200
    assert "user_id" in response.json()
```

### Using Fixtures
```python
def test_with_test_user(test_user):
    """Test using pre-created user"""
    assert test_user["user_id"] is not None
    assert test_user["email"] is not None
```

### Parametrized Tests
```python
@pytest.mark.parametrize("email,expected", [
    ("valid@example.com", 200),
    ("invalid-email", 400),
    ("", 400),
])
def test_email_validation(api_client, email, expected):
    response = api_client.post("/api/v1/users", json={"email": email})
    assert response.status_code == expected
```

## 🔧 Fixtures

### Available Fixtures

- `api_client` - HTTP client for making requests
- `test_data` - Factory for generating test data
- `test_user` - Pre-created test user (auto-cleanup)
- `authenticated_client` - API client with auth token

### Custom Fixtures

Add to `conftest.py`:
```python
@pytest.fixture
def custom_fixture():
    # Setup
    yield data
    # Teardown
```

## 📈 CI/CD Integration

### GitHub Actions
```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Python
        uses: actions/setup-python@v2
        with:
          python-version: '3.11'
      - name: Install dependencies
        run: |
          cd python-pytest
          pip install -r requirements.txt
      - name: Run tests
        run: |
          cd python-pytest
          pytest --junitxml=reports/junit.xml
      - name: Publish Test Results
        uses: EnricoMi/publish-unit-test-result-action@v2
        if: always()
        with:
          files: python-pytest/reports/junit.xml
```

## 🐛 Debugging

### Run with Debugging
```bash
pytest --pdb  # Drop into debugger on failure
```

### Show Locals on Failure
```bash
pytest -l
```

### Stop on First Failure
```bash
pytest -x
```

### Run Last Failed Tests
```bash
pytest --lf
```

## 📊 Reports

Reports are generated in `reports/` directory:
- `report.html` - HTML test report
- `junit.xml` - JUnit XML for CI/CD
- `coverage/` - Coverage reports

## 🎯 Best Practices

1. **Independent Tests** - Each test should be self-contained
2. **Use Fixtures** - For setup and teardown
3. **Descriptive Names** - Clear test names
4. **Assertions** - Use meaningful assertions
5. **Cleanup** - Always cleanup test data
6. **Markers** - Tag tests appropriately
7. **Parametrize** - For testing multiple scenarios

## 🔍 Troubleshooting

### Tests Failing
1. Check API is running: `curl http://localhost:8080`
2. Verify `.env` configuration
3. Check test data validity
4. Review error messages

### Connection Errors
1. Ensure API Gateway is running
2. Check `API_BASE_URL` in `.env`
3. Verify network connectivity

### Authentication Errors
1. Check credentials in `.env`
2. Verify token generation
3. Check token expiration

## 📚 Resources

- [pytest Documentation](https://docs.pytest.org/)
- [requests Documentation](https://requests.readthedocs.io/)
- [Faker Documentation](https://faker.readthedocs.io/)

## 🤝 Contributing

1. Add tests in appropriate file
2. Follow naming conventions
3. Add markers
4. Update documentation
5. Run all tests before committing

## 📝 Example Commands

```bash
# Quick smoke test
pytest -m smoke -v

# Full regression
pytest -m regression

# Parallel execution with coverage
pytest -n 4 --cov=. --cov-report=html

# Generate report
pytest --html=reports/report.html --self-contained-html

# Run specific tests
pytest tests/test_users.py::TestUserManagement::test_create_user_success -v
```

Happy Testing! 🎉
