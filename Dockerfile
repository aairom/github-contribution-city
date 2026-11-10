# Multi-stage optimized Dockerfile/Containerfile for GitHub Contribution City

# Stage 1: Build the Go binary statically
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install CA certificates for HTTPS requests to GitHub API
RUN apk add --no-cache ca-certificates git

# Cache Go modules
COPY go.mod ./
RUN go mod download

# Copy source files
COPY . .

# Build statically linked binary with stripped debug symbols for minimum size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/github-contribution-city cmd/city/main.go

# Stage 2: Minimal runtime image (distroless / scratch / alpine)
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/bin/github-contribution-city /usr/local/bin/github-contribution-city
COPY --from=builder /app/.env.example /app/.env.example

# Create input and output directories
RUN mkdir -p /app/input /app/output && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8088

ENV PORT=8088 \
    HOST=0.0.0.0 \
    OUTPUT_DIR=/app/output \
    INPUT_DIR=/app/input

ENTRYPOINT ["github-contribution-city"]
CMD ["-server"]
