FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o postgres-sidecar .

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/postgres-sidecar .
COPY --from=builder /app/config ./config
COPY --from=builder /app/auth ./auth
COPY --from=builder /app/handlers ./handlers

EXPOSE 8080
CMD ["./postgres-sidecar"]
