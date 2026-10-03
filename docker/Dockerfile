FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
