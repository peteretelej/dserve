#!/bin/bash
set -e

goreleaser --help 2>&1 >/dev/null
if [ $? -gt 0 ]
then
	echo "FAILED: goreleaser binary not in path, you should:"
	echo -e "\tgo get -u github.com/goreleaser/goreleaser"
	exit 1
fi

go vet ./...

rm -rf dist/*

goreleaser

cd ../
echo "================= dist DONE."
