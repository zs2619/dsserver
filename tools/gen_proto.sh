#!/bin/bash
SRC_DIR=../proto
DST_DIR=../pb
export PATH=$PATH:`pwd`/protobuf/bin
chmod +x `pwd`/protobuf/bin/protoc
chmod +x `pwd`/protobuf/bin/protoc-gen-go
chmod +x `pwd`/protobuf/bin/protoc-gen-go-grpc
protoc --go_opt=paths=source_relative  --go-grpc_opt=paths=source_relative --proto_path=$SRC_DIR  --go-grpc_out=$DST_DIR --go_out=$DST_DIR  $SRC_DIR/*.proto