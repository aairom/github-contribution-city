# Multi-stage optimized Containerfile (Podman / Docker compatible)
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache ca-certificates git

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/github-contribution-city cmd/city/main.go

FROM alpine:3.24

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/bin/github-contribution-city /usr/local/bin/github-contribution-city
COPY --from=builder /app/.env.example /app/.env.example

RUN mkdir -p /app/input /app/output && chown -R appuser:appgroup /app
USER appuser

EXPOSE 8088

ENV PORT=8088 \
    HOST=0.0.0.0 \
    OUTPUT_DIR=/app/output \
    INPUT_DIR=/app/input

ENTRYPOINT ["github-contribution-city"]
CMD ["-server"]
