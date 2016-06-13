#!/bin/bash

PROJECT_DIR=`dirname $0`
echo `cd $PROJECT_DIR;pwd`

# GOPATH=/Users/tokusei/.go GODEBUG=cgocheck=0 go run serial.go -port /dev/cu.usbserial

# build mac
mkdir -p target/macosx
GOPATH=/Users/tokusei/.go GODEBUG=cgocheck=0 go build -o target/macosx/robip-tool

