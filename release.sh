#!/bin/bash
set -e 

# confirm goreleaser exists
goreleaser --help 2>&1 >/dev/null

go vet ./...

rm -rf dist/*

goreleaser

cd ../
echo "================= dist DONE."
