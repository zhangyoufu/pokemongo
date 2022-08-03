#!/bin/sh
cd -- "$(dirname -- "$0")"
DIR=/home/yongheng/challs/committing/ne0_06_15_bibi_pwn/pokemongo-pwn-docker/源码
docker run --rm -v "$(pwd)":"$DIR" -v "$(pwd)/git":/usr/local/bin/git:ro -w "$DIR" -i golang:1.18.0 <<EOF
set -ex
go build .
strip pokemongo
objcopy --update-section .note.go.buildid=go-buildid pokemongo
EOF
