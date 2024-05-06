FROM golang:latest AS builder

WORKDIR /app

COPY webserver .

RUN go mod download
RUN go build -o app .

WORKDIR /app

# Install required libraries for GLIBC 2.32
RUN apt-get update && apt-get install -y --no-install-recommends libc6

COPY kubernetes/nextcloud.yml .

CMD ["./app"]