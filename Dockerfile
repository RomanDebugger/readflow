FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o readflow-engine ./src/main.go

FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/readflow-engine .
RUN mkdir -p ./data/input_pdfs ./data/chunks
EXPOSE 8080
CMD ["./readflow-engine"]