FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy service and local module go.mod files to satisfy local replaces
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY api/pb/userpb/go.mod ./api/pb/userpb/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY api/pb/departmentpb/go.mod ./api/pb/departmentpb/
COPY api/pb/designationpb/go.mod ./api/pb/designationpb/
COPY pkg/middleware/go.mod ./pkg/middleware/

# Download dependencies for the user-service module
RUN cd services/user-service && go mod download

# Copy source code
COPY . .

# Build the service from within the module directory
RUN cd services/user-service && CGO_ENABLED=0 GOOS=linux go build -o /app/user-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/user-service .

EXPOSE 50051

CMD ["./user-service"]
