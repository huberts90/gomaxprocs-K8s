FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY ./ ./
RUN CGO_ENABLED=0 go build -o app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"] 