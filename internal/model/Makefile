.PHONY: proto

all: proto

# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
proto:
	protoc -I. -I="${GOPATH}/src" -I="${GOPATH}/src/github.com/gogo/protobuf/protobuf" --gogofaster_out=. *.proto
