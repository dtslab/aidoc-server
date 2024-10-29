# Dockerfile
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o main ./cmd

# Runtime image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Expose the port your application listens on
EXPOSE 8080

CMD ["./main"]