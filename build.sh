#!/bin/bash
set -e

go vet ./...

rm -rf dist/*

# windows and linux x64
echo "Building Linuxamd64, Windowsx64"
GOOS=linux CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "static"'  -o dist/dserve main.go && \
GOOS=windows CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "static"'  -o dist/dserve.exe main.go
cd dist/
zip -r dserve_for_windowsx64.zip dserve.exe && \
tar -cvzf dserve_for_linux_amd64.tar.gz dserve && \
rm dserve{,.exe}

cd ../
# Windows 32bit
echo "Building Windows, Linux 32 bit"
GOARCH=386 GOOS=linux CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "static"'  -o dist/dserve main.go && \
GOARCH=386 GOOS=windows CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "static"'  -o dist/dserve.exe main.go && \
cd dist/
zip -r dserve_for_windows.zip dserve.exe && \
tar -cvzf dserve_for_linux.tar.gz dserve && \
rm dserve{,.exe}

cd ../
# Mac, iOS 386 (darwin/386)
echo "Building for Mac"
GOARCH=386 GOOS=darwin CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "static"'  -o dist/dserve main.go && \
cd dist/
zip -r dserve_for_mac.zip dserve && \
rm dserve

cd ../
echo "================= Releases DONE, see dist/."
