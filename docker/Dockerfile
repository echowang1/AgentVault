FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY server/go.mod ./server/go.mod
RUN cd server && go mod download

COPY server ./server
RUN CGO_ENABLED=0 go build -o /app/server ./server/cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/server ./server
COPY docker/entrypoint.sh ./entrypoint.sh

RUN chmod +x /app/entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
