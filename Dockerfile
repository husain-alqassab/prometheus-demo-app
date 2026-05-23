# Multi-stage build for minimal image size
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o prometheus-demo-app .

# Final stage - use distroless for minimal, secure image
FROM gcr.io/distroless/static-debian12

# Copy the binary from builder
COPY --from=builder /app/prometheus-demo-app /prometheus-demo-app

# Use non-root user (distroless runs as nonroot by default)
# Expose port 8080
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/prometheus-demo-app"]