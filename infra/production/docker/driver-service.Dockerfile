FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/driver-service/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o driver-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/driver-service/cmd/driver-service .
CMD ["./driver-service"] 