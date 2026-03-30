# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o readflow-engine ./src/main.go

# Stage 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/readflow-engine .
COPY --from=builder /app/data ./data
EXPOSE 8080
CMD ["./readflow-engine"]