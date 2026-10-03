FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server && go build -o purge ./cmd/purge

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget dcron
WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/purge ./purge
COPY --from=builder /app/migrations ./migrations
COPY docker/purge-crontab /etc/crontabs/root
RUN chmod 0644 /etc/crontabs/root
EXPOSE 8080
CMD ["./server"]
