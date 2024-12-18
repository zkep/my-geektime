FROM golang:1.23 AS builder

ENV GOPROXY="https://goproxy.cn,direct"

WORKDIR /app
COPY . /app

RUN make all

FROM jrottenberg/ffmpeg:7.1-ubuntu2404

RUN apt update && apt install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/mygeektime /usr/bin/mygeektime

EXPOSE 8090

ENTRYPOINT ["mygeektime"]
