# Step 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Cache dependencies
RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  go mod download -x


# Build the Go app
RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=bind,target=. \
  go build -o /build/myapp

# Step 2: Create a smaller image and copy the binary
FROM alpine:3.20

# Install Chromium

RUN apk add --no-cache chromium
RUN addgroup -S myuser && adduser -S myuser -G myuser

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/myapp .

# Use an unprivileged user
USER myuser

ENV MODE=docker

# Command to run the executable
ENTRYPOINT ["./myapp", "run"]
