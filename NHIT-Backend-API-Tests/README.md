# NHIT Backend API Testing Suite

## 🎯 Overview

Comprehensive automated testing suite for NHIT Backend REST APIs using multiple testing frameworks and tools.

## 📁 Project Structure

```
NHIT-Backend-API-Tests/
├── postman/                    # Postman collections
│   ├── collections/
│   │   ├── user-management.json
│   │   └── authentication.json
│   ├── environments/
│   │   ├── local.json
│   │   ├── dev.json
│   │   └── prod.json
│   └── newman-reports/
│
├── python-pytest/              # Python pytest tests
│   ├── tests/
│   │   ├── test_users.py
│   │   ├── test_auth.py
│   │   └── test_integration.py
│   ├── utils/
│   │   ├── api_client.py
│   │   └── test_data.py
│   ├── requirements.txt
│   └── pytest.ini
│
├── javascript-jest/            # JavaScript/Jest tests
│   ├── tests/
│   │   ├── users.test.js
│   │   └── auth.test.js
│   ├── utils/
│   │   └── apiClient.js
│   ├── package.json
│   └── jest.config.js
│
├── k6/                         # Load testing with k6
│   ├── load-tests/
│   │   ├── user-load.js
│   │   └── auth-load.js
│   └── scenarios/
│
├── playwright/                 # E2E API testing
│   ├── tests/
│   │   └── api.spec.ts
│   └── playwright.config.ts
│
├── docker-compose.yml          # Run all tests in containers
├── .env.example
└── README.md
```

## 🚀 Quick Start

### Option 1: Python pytest (Recommended)
```bash
cd python-pytest
pip install -r requirements.txt
pytest -v
```

### Option 2: JavaScript/Jest
```bash
cd javascript-jest
npm install
npm test
```

### Option 3: Postman/Newman
```bash
cd postman
newman run collections/user-management.json -e environments/local.json
```

### Option 4: k6 Load Testing
```bash
cd k6
k6 run load-tests/user-load.js
```

## 📊 Test Coverage

- ✅ User Management APIs (CRUD)
- ✅ Authentication APIs (Login, Register, Logout)
- ✅ Role Management
- ✅ Error Handling
- ✅ Validation Tests
- ✅ Integration Tests
- ✅ Load/Performance Tests

## 🎯 Features

- **Multiple Frameworks** - Choose your preferred testing tool
- **CI/CD Ready** - GitHub Actions, GitLab CI configurations
- **Parallel Execution** - Fast test runs
- **Detailed Reports** - HTML, JSON, JUnit formats
- **Environment Management** - Local, Dev, Staging, Prod
- **Test Data Management** - Fixtures and factories
- **Mocking Support** - For external dependencies

## 📝 Test Types

### 1. Unit Tests
- Individual endpoint testing
- Request/response validation
- Error handling

### 2. Integration Tests
- Multi-step workflows
- Cross-service interactions
- Data consistency

### 3. Load Tests
- Performance benchmarks
- Stress testing
- Scalability validation

### 4. Security Tests
- Authentication validation
- Authorization checks
- Input sanitization

## 🔧 Configuration

Edit `.env` file:
```env
API_BASE_URL=http://localhost:8080
API_TIMEOUT=30000
TEST_USER_EMAIL=test@example.com
TEST_USER_PASSWORD=testpass123
```

## 📈 Running Tests

### Run All Tests
```bash
# Python
pytest

# JavaScript
npm test

# Postman
newman run postman/collections/*.json
```

### Run Specific Test Suite
```bash
# Python
pytest tests/test_users.py -v

# JavaScript
npm test -- users.test.js
```

### Generate Reports
```bash
# Python with HTML report
pytest --html=reports/report.html

# JavaScript with coverage
npm test -- --coverage
```

## 🐳 Docker Support

Run tests in containers:
```bash
docker-compose up --build
```

## 📊 CI/CD Integration

### GitHub Actions
```yaml
- name: Run API Tests
  run: |
    cd python-pytest
    pytest --junitxml=reports/junit.xml
```

### GitLab CI
```yaml
test:
  script:
    - cd python-pytest
    - pytest -v
```

## 🎉 Best Practices

1. **Independent Tests** - Each test should be self-contained
2. **Clean State** - Reset data between tests
3. **Meaningful Names** - Descriptive test names
4. **Fast Execution** - Optimize for speed
5. **Reliable** - No flaky tests

## 📚 Documentation

- [Python pytest Guide](python-pytest/README.md)
- [JavaScript Jest Guide](javascript-jest/README.md)
- [Postman Guide](postman/README.md)
- [k6 Load Testing Guide](k6/README.md)

## 🤝 Contributing

1. Add new tests in appropriate directory
2. Follow naming conventions
3. Update documentation
4. Run all tests before committing

## 📞 Support

For issues or questions, refer to the main project documentation.
