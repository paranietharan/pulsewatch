# Multi-stage build for pulse-watch-db migration runner
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pulsewatch-migrator .

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 10001 appgroup && \
    adduser -S -u 10001 -G appgroup appuser

COPY --from=builder /app/bin/pulsewatch-migrator /app/pulsewatch-migrator
COPY --from=builder /app/migrations /app/migrations

USER 10001:10001

ENTRYPOINT ["/app/pulsewatch-migrator"]
CMD ["up"]
