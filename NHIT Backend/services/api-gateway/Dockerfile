FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy service and local module go.mod files
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY api/proto/go.mod api/proto/go.sum ./api/proto/
COPY api/pb/userpb/go.mod ./api/pb/userpb/
COPY api/pb/departmentpb/go.mod ./api/pb/departmentpb/
COPY api/pb/designationpb/go.mod ./api/pb/designationpb/
COPY api/pb/organizationpb/go.mod ./api/pb/organizationpb/
COPY api/pb/vendorpb/go.mod ./api/pb/vendorpb/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY api/pb/projectpb/go.mod ./api/pb/projectpb/

# Download dependencies for the api-gateway module
RUN cd services/api-gateway && go mod download

# Copy source code
COPY . .

# Build the gateway from within the module directory
RUN cd services/api-gateway && CGO_ENABLED=0 GOOS=linux go build -o /app/api-gateway ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
