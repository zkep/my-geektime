FROM golang:1.23 AS builder

ENV GOPROXY="https://goproxy.cn,direct"

WORKDIR /app
COPY . /app

RUN make all

FROM alpine:latest

RUN apk add --no-cache tzdata

COPY --from=builder /app/mygeektime /usr/bin/

EXPOSE 8090
