# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/app ./cmd/app

# Stage 2: Runtime
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/app .

# Fallback config jika volume belum di-mount
COPY configs/ ./configs/
COPY storage/ ./storage/

EXPOSE 3001

CMD ["./app"]