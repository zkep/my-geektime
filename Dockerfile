FROM golang:1.23 AS builder

ENV GOPROXY="https://goproxy.cn,direct"

WORKDIR /app
COPY . /app

RUN make all

FROM ubuntu:22.04

RUN apt-get update -y  && apt-get upgrade -y && apt-get install wget
RUN wget https://ffmpeg.org/releases/ffmpeg-7.1.tar.xz
RUN tar -xvf ffmpeg-7.1.tar.xz
RUN cd ffmpeg-7.1.tar.xz
RUN ./configure --enable-gpl --enable-libx264
RUN make && make  install

COPY --from=builder /app/mygeektime /usr/bin/

EXPOSE 8090
