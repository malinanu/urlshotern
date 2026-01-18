# Multi-stage build for production
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata postgresql-client

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o url-shortener cmd/server/main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies including PostgreSQL client
RUN apk --no-cache add ca-certificates tzdata postgresql-client bash

# Create app user
RUN adduser -D -s /bin/sh appuser

WORKDIR /home/appuser

# Copy binary from builder stage
COPY --from=builder /app/url-shortener .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Copy migration files and scripts
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/configs/migrations ./configs/migrations
COPY --from=builder /app/migrations ./migrations

# Fix script permissions and line endings
RUN find ./scripts -name "*.sh" -exec sed -i 's/\r$//' {} \; && \
    chmod +x ./scripts/*.sh

# Copy entrypoint script
COPY docker/entrypoint.sh /entrypoint.sh
RUN sed -i 's/\r$//' /entrypoint.sh && chmod +x /entrypoint.sh

# Change ownership to appuser
RUN chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use entrypoint to run migrations before starting app
ENTRYPOINT ["/entrypoint.sh"]

# Run the binary
CMD ["./url-shortener"]
