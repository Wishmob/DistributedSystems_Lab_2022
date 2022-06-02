#!/bin/sh
(cd cloud_server || exit
mkdir proto
protoc --go-grpc_out=proto --proto_path=../ ../proto.proto
protoc --go_out=proto --proto_path=../ ../proto.proto)

(cd database || exit
mkdir proto
protoc --go-grpc_out=proto --proto_path=../ ../proto.proto
protoc --go_out=proto --proto_path=../ ../proto.proto)