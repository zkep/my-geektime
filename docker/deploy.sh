#!/bin/bash
PROJ=`pwd | awk -F'/' '{print $(NF)}'`

echo "go build"
go mod tidy


env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags=jsoniter -o $PROJ main.go

chmod +x "$PROJ"

echo "打包"
tar -zcvf "$PROJ".tar.gz \
--no-xattrs \
--exclude '.DS_Store' \
--exclude 'logs/' \
--exclude '__MACOSX' \
"$PROJ"

scp "$PROJ".tar.gz root@8.141.6.243:/root/go/bin