FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy service and local module go.mod files
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/
COPY api/proto/go.mod api/proto/go.sum ./api/proto/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY api/pb/organizationpb/go.mod ./api/pb/organizationpb/
COPY api/pb/userpb/go.mod ./api/pb/userpb/

# Download dependencies for the auth-service module
RUN cd services/auth-service && go mod download

# Copy source code
COPY . .

# Build the service from within the module directory
RUN cd services/auth-service && CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth-service .

EXPOSE 50052

CMD ["./auth-service"]
