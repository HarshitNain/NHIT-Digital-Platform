# Multi-stage build for greennote-service

FROM golang:1.22 AS builder
WORKDIR /app

# Pre-cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source
COPY . .

# Build the greennote-service binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /greennote \
    ./services/greennote-service/cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /

COPY --from=builder /greennote /greennote

EXPOSE 50051 8080

ENTRYPOINT ["/greennote"]
