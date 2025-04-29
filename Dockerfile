FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install git for version information
RUN apk add --no-cache git

# Set build arguments with defaults
ARG VERSION=dev
ARG BUILD_TIME

# Set default build time if not provided
RUN if [ -z "$BUILD_TIME" ]; then BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ"); fi && \
    echo "Building version: $VERSION, build time: $BUILD_TIME" && \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/samzong/mdctl/cmd.Version=${VERSION} -X github.com/samzong/mdctl/cmd.BuildTime=${BUILD_TIME}" -o /app/bin/mdctl

# Use a minimal alpine image for the final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/bin/mdctl /usr/local/bin/mdctl

# Create config directory
RUN mkdir -p /root/.config/mdctl

# Set the entrypoint
ENTRYPOINT ["mdctl"]

# Default command
CMD ["--help"]
