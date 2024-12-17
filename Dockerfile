FROM golang:1.23 AS builder

ENV GOPROXY="https://goproxy.cn,direct"

WORKDIR /app
COPY . /app

RUN make all

FROM ubuntu:22.04

RUN apt update -y --fix-missing
RUN apt install git -y --fix-missing
RUN git clone https://github.com/FFmpeg/FFmpeg.git
RUN cd FFmpeg
RUN ./configure --enable-gpl --enable-libx264
RUN make && make  install

COPY --from=builder /app/mygeektime /usr/bin/

EXPOSE 8090
