#!/bin/bash

set -xe
go test ./...
go build -o empired ./service
./empired --fs-root $PWD/example/static --base etc/empire
./empired --fs-root $PWD/example/git --base etc/empire
