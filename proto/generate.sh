#!/bin/bash

# Generate all ProtoMessages.

protoc --go_out=. --go-grpc_out=. *.proto