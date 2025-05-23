FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o sidecar .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sidecar .

EXPOSE 8080
CMD ["./sidecar"]
