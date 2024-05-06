FROM golang:latest AS builder

WORKDIR /app

COPY webserver .

RUN go mod download
RUN go build -o app .

FROM debian:buster-slim

WORKDIR /app

COPY --from=builder /app/app .
COPY kubernetes/nextcloud.yml .

CMD ["./app"]