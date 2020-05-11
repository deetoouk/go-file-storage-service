#!/bin/bash

SRC_DIR=filestorage
DST_DIR=$SRC_DIR

#protoc -I=$SRC_DIR --go_opt=paths=source_relative --go_out=$DST_DIR $SRC_DIR/fileservice.proto

# protoc -I $SRC_DIR/ $SRC_DIR/fileservice.proto --go_out=plugins=grpc:helloworld

# protoc --proto_path=$SRC_DIR --go_out=$DST_DIR --go_opt=paths=source_relative $SRC_DIR/fileservice.proto

protoc -I $SRC_DIR/ $SRC_DIR/filestorage.proto --go_opt=paths=source_relative --go_out=plugins=grpc:filestorage
