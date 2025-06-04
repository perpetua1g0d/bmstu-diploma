FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o postgres-sidecar .

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/postgres-sidecar .
COPY --from=builder /app/config ./config
COPY --from=builder /app/handlers ./handlers

COPY postgres-setup.sh /postgres-setup.sh
RUN chmod +x /postgres-setup.sh
EXPOSE 8080
ENTRYPOINT ["/postgres-setup.sh"]

# EXPOSE 8080
# CMD ["./postgres-sidecar"]
