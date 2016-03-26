#!/bin/bash
rm -rf releases/*
GOOS=linux go build -o releases/dserve main.go && \
GOOS=windows go build -o releases/dserve.exe main.go && \
echo "Windows and Linux executable binaries built. Located in releases folder." && \
zip -r releases/dserve_for_windows.zip releases/dserve.exe && \
tar -cvzf releases/dserve_for_linux.tar.gz releases/dserve && \
echo "Release builds completed"
