# Use the official golang image to create a binary.
FROM golang:1.23-bookworm AS builder

# Set environment variables for static build
ENV CGO_ENABLED=0 GOOS=linux

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
COPY go.* ./
RUN go mod download all

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -v -o main

# Use the official Debian slim image for a lean production container.
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/main /app/main

# Set working directory
WORKDIR /app

# Expose application port
EXPOSE 8080

# Run the web service on container startup.
CMD ["./main"]