# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Jangan build saat build image — biar runtime-nya go run
CMD ["go", "run", "api/index.go"]
