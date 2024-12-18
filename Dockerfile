FROM golang:1.23 AS builder

ENV GOPROXY="https://goproxy.cn,direct"

WORKDIR /app
COPY . /app

RUN make all

FROM jrottenberg/ffmpeg

RUN apt update -y
RUN apt install libc6 -y

COPY --from=builder /app/mygeektime /usr/bin/mygeektime

EXPOSE 8090

ENTRYPOINT ["mygeektime"]
