# Multi-stage build for pulsewatch-notification-service
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pulsewatch-notification-service .

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 10001 appgroup && \
    adduser -S -u 10001 -G appgroup appuser

COPY --from=builder /app/bin/pulsewatch-notification-service /app/pulsewatch-notification-service

USER 10001:10001

ENTRYPOINT ["/app/pulsewatch-notification-service"]
