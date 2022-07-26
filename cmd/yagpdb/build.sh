#!/bin/sh
#VERSION=$(git describe --tags)
VERSION=$(git describe --tags | awk -F "-" '{print $1}')
echo Building version $VERSION
go build -ldflags "-X github.com/botlabs-gg/yagpdb/common.VERSION=${VERSION}"
