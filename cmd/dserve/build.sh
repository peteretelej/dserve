#!/bin/bash
set -e

rm -rf releases/*

# windows and linux x64
echo "Building Linuxamd64, Windowsx64"
GOOS=linux go build -o releases/dserve main.go && \
GOOS=windows go build -o releases/dserve.exe main.go && \
zip -r releases/dserve_for_windowsx64.zip releases/dserve.exe && \
tar -cvzf releases/dserve_for_linux_amd64.tar.gz releases/dserve && \
rm releases/dserve{,.exe}

# Windows 32bit
echo "Building Windows, Linux 32 bit"
GOARCH=386 GOOS=linux go build -o releases/dserve main.go && \
GOARCH=386 GOOS=windows go build -o releases/dserve.exe main.go && \
zip -r releases/dserve_for_windows.zip releases/dserve.exe && \
tar -cvzf releases/dserve_for_linux.tar.gz releases/dserve && \
rm releases/dserve{,.exe}

# Mac, iOS 386 (darwin/386)
echo "Building for Mac"
GOARCH=386 GOOS=darwin go build -o releases/dserve main.go && \
zip -r releases/dserve_for_mac.zip releases/dserve && \
rm releases/dserve

echo "BUILDS completed. ================="
