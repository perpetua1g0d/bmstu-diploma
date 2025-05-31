FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o idp .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/idp .
EXPOSE 8080
CMD ["./idp"]
