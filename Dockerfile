# --- Builder Stage ---
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install C build tools needed for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev

# Copy module files and download dependencies first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the main server application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o go-submission-service .

# Build the CLI tool (optional)
# RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o cli ./cmd/cli

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# Copy the built binaries from the builder stage
COPY --from=builder /app/go-submission-service .
# COPY --from=builder /app/cli .

# Copy templates
COPY templates ./templates

# Expose the port the application runs on
EXPOSE 8080

# Command to run the executable
CMD ["./go-submission-service"] 