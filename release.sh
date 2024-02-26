#!/bin/bash
LastTag=$(git tag -l --sort=-version:refname "v*" | head -n 1)
OS=("openbsd" "freebsd" "darwin" "linux" "windows")
ARCH=("amd64" "386")

for os in ${OS[@]}
do
	echo "Building $os..."
	for arch in ${ARCH[@]}
	do
		echo "+$arch"
		if [ "$os" = "windows" ]; then
			ext=".exe"
		fi
		GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o "toonily-dl-$LastTag-$os-$arch$ext"
	done
done
