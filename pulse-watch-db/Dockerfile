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

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/pulsewatch-migrator /app/pulsewatch-migrator
COPY --from=builder /app/migrations /app/migrations

ENTRYPOINT ["/app/pulsewatch-migrator"]
CMD ["up"]
