# Multi-stage build for pulsewatch-probe-worker
FROM golang:1.27.3-alpine3.24 AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pulsewatch-probe-worker ./cmd/worker

FROM alpine:3.24

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 10001 appgroup && \
    adduser -S -u 10001 -G appgroup appuser

COPY --from=builder /app/bin/pulsewatch-probe-worker /app/pulsewatch-probe-worker

USER 10001:10001

ENTRYPOINT ["/app/pulsewatch-probe-worker"]
