FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy service and local module go.mod files
COPY services/vendor-service/go.mod services/vendor-service/go.sum ./services/vendor-service/
COPY api/pb/vendorpb/go.mod ./api/pb/vendorpb/
COPY api/pb/authpb/go.mod ./api/pb/authpb/
COPY api/pb/projectpb/go.mod ./api/pb/projectpb/
COPY api/pb/organizationpb/go.mod ./api/pb/organizationpb/
COPY pkg/middleware/go.mod ./pkg/middleware/

# Download dependencies
RUN cd services/vendor-service && go mod download

# Copy source code
COPY . .

# Build the service
RUN cd services/vendor-service && CGO_ENABLED=0 GOOS=linux go build -o /app/vendor-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/vendor-service .

EXPOSE 50058 8084

CMD ["./vendor-service"]
