# syntax=docker/dockerfile:1

ARG APP_NAME=leru

FROM golang:1.26 AS builder
ARG APP_NAME
WORKDIR /src

# Cache dependencies
COPY go.mod ./
RUN go mod download

# Copy source and build
COPY . .
RUN mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/${APP_NAME}

FROM alpine:3.18
ARG APP_NAME
ENV APP_NAME=${APP_NAME}

RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /src/bin/${APP_NAME} /app/${APP_NAME}

EXPOSE 9090
ENTRYPOINT ["/bin/sh", "-c", "/app/${APP_NAME}"]
