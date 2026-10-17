FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY services/department-service/go.mod services/department-service/go.sum ./services/department-service/
COPY api/pb/departmentpb/go.mod ./api/pb/departmentpb/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY pkg/middleware/go.mod ./pkg/middleware/

# Download dependencies
RUN cd services/department-service && go mod download

# Copy source code
COPY . .

# Build the service
RUN cd services/department-service && CGO_ENABLED=0 GOOS=linux go build -o /department-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /department-service .

EXPOSE 50054

CMD ["./department-service"]
