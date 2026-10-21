# Multi-stage build for pulsewatch-api
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pulsewatch-api ./cmd/api

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -S -g 10001 appgroup && \
    adduser -S -u 10001 -G appgroup appuser

COPY --from=builder /app/bin/pulsewatch-api /app/pulsewatch-api

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/pulsewatch-api"]
