FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy service and local module go.mod files to satisfy local replaces
COPY services/organization-service/go.mod services/organization-service/go.sum ./services/organization-service/
COPY api/pb/organizationpb/go.mod ./api/pb/organizationpb/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY api/pb/projectpb/go.mod ./api/pb/projectpb/
COPY pkg/middleware/go.mod ./pkg/middleware/
COPY services/shared/go.mod services/shared/go.sum ./services/shared/

# Download dependencies for the organization-service module
RUN cd services/organization-service && go mod download

# Copy source code
COPY . .

# Build the service from within the module directory
RUN cd services/organization-service && CGO_ENABLED=0 GOOS=linux go build -o /app/organization-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/organization-service .

EXPOSE 50053

CMD ["./organization-service"]
