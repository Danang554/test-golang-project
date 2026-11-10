# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Install Goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy source code
COPY . .

# Compile statically linked binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/seeder ./cmd/seeder

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install standard CA certificates and tzdata
RUN apk --no-cache add ca-certificates tzdata

# Copy binaries and assets
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/bin/api /app/bin/api
COPY --from=builder /app/bin/seeder /app/bin/seeder
COPY database/migrations /app/database/migrations
COPY .env .env

# Expose HTTP port
EXPOSE 8080

# Run server
CMD ["/app/bin/api"]
