# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY services/designation-service/go.mod services/designation-service/go.sum ./services/designation-service/
COPY api/pb/designationpb/go.mod ./api/pb/designationpb/

# Download dependencies
RUN cd services/designation-service && go mod download

# Copy source code
COPY . .

# Build the service
RUN cd services/designation-service && CGO_ENABLED=0 GOOS=linux go build -o /app/designation-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/designation-service .

EXPOSE 50055

CMD ["./designation-service"]
