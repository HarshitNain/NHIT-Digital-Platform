FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy test files
COPY . .

# Create reports directory
RUN mkdir -p reports

# Run tests
CMD ["pytest", "-v"]
