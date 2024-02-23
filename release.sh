#!/bin/bash
LastTag=$(git tag --list|head -n 1)
OS=("plan9" "openbsd" "freebsd" "darwin" "linux" "windows")
ARCH=("amd64" "386")

for os in ${OS[@]}
do
	echo "Building $os..."
	for arch in ${ARCH[@]}
	do
		echo "+$arch"
		GOOS="$OS" GOARCH="$arch" go build -o "toonily-dl-$LastTag-$os-$arch"
	done
done