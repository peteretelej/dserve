#!/bin/bash
rm -rf releases/*
GOOS=linux go build -o releases/dserve main.go && \
GOOS=windows go build -o releases/dserve.exe main.go && \
echo "Windows and Linux executable binaries built. Located in releases folder."
zip -r releases/dserve_for_windowsx64.zip releases/dserve.exe && \
tar -cvzf releases/dserve_for_linux_amd64.tar.gz releases/dserve && \
rm releases/dserve releases/dserve.exe
rm -rf releases/dserve.exe
GOARCH=386 GOOS=windows go build -o releases/dserve.exe main.go && \
echo "Windows 32bit executable binaries built. Located in releases folder."
zip -r releases/dserve_for_windows.zip releases/dserve.exe && \
rm releases/dserve releases/dserve.exe
echo "Release builds completed"
