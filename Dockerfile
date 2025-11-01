# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files and vendored dependencies
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy source code
COPY . .

# Build the application using vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o main .

# Runtime stage - use distroless for minimal image
FROM gcr.io/distroless/static-debian11

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port 8080
EXPOSE 8080

# Set default port
ENV PORT=8080

# Run the application
CMD ["./main"]
