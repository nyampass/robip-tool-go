#!/bin/bash

GOPATH=/Users/tokusei/.go ~/.go/bin/gox -osarch="darwin/amd64" ./...

mv robip-tool-go_darwin_amd64 "./target/macosx/Robip Tool.app/Contents/MacOS/robip-tool-go"

rm "./target/macosx/Robip Tool.zip"
zip -r "./target/macosx/Robip Tool.zip" "./target/macosx/Robip Tool.app"

# GOPATH=/Users/tokusei/.go ~/.go/bin/gox -osarch="windows/amd64" ./...
