# Multi-stage build for pulsewatch-probe-worker
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pulsewatch-probe-worker ./cmd/worker

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/pulsewatch-probe-worker /app/pulsewatch-probe-worker

ENTRYPOINT ["/app/pulsewatch-probe-worker"]
