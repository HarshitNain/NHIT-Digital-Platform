# ==============================================================================
# NHIT ENTERPRISE API GATEWAY - PRODUCTION DOCKERFILE
# FastAPI / PyJWT / RBAC / Multi-Tenant Gateway
# ==============================================================================

FROM python:3.12-slim

WORKDIR /app

# Install system dependencies for psycopg2 and health checks
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy backend application code
COPY . .

# Environment defaults (overridden by docker-compose or .env)
ENV PYTHONUNBUFFERED=1 \
    PORT=8080 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_NAME=nhit_db \
    DB_USER=postgres \
    DB_PASSWORD=shristi

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/health || exit 1

CMD ["python", "gateway_swagger.py"]
