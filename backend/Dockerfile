# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/app ./cmd/app

# Stage 2: Runtime
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata python3 py3-pip py3-numpy py3-pandas py3-matplotlib && \
    pip3 install --break-system-packages --no-cache-dir mplfinance ta && \
    ln -sf python3 /usr/bin/python

COPY --from=builder /app/bin/app .
COPY --from=builder /app/scripts ./scripts

COPY configs/ ./configs/
COPY storage/ ./storage/

EXPOSE 8001

CMD ["./app"]
