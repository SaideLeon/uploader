# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
# -ldflags="-w -s" reduces the size of the binary by stripping debugging information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/uploader .

# Stage 2: Create the final, lightweight image
FROM alpine:latest

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/uploader .

# Create the uploads directory and set ownership BEFORE switching to appuser
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app/uploads

# Switch to the non-root user
USER appuser

# Expose port 8002 to the outside world
EXPOSE 8002

# Command to run the executable
CMD ["./uploader"]