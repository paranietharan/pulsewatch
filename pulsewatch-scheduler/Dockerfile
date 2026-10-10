# Multi-stage build for pulsewatch-scheduler
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pulsewatch-scheduler ./cmd/scheduler

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/pulsewatch-scheduler /app/pulsewatch-scheduler

ENTRYPOINT ["/app/pulsewatch-scheduler"]
