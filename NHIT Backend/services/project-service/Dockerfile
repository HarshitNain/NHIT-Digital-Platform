# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY services/project-service/go.mod services/project-service/go.sum ./services/project-service/
COPY api/pb/projectpb/go.mod ./api/pb/projectpb/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY pkg/middleware/go.mod ./pkg/middleware/

# Download dependencies
RUN cd services/project-service && go mod download

# Copy source code
COPY . .

# Build the application
RUN cd services/project-service && CGO_ENABLED=0 GOOS=linux go build -o /app/project-service ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/project-service .

# Expose ports
EXPOSE 50057 8057

# Run the binary
CMD ["./project-service"]
