#!/bin/sh

DIFF=`git diff --stat`
VERSION=`git describe --tags --always --dirty`
NOW=`date`
COMMIT=`git rev-parse --short HEAD`

if test ! -z "$DIFF"; then
  STATE='dirty'
else
  STATE='clean'
fi

FLAGS="-X 'main.gitCommit=$COMMIT' -X 'main.buildTime=$NOW' -X 'main.treeState=$STATE' -X 'main.tag=$VERSION' -w -s"

if [ ! -f ./go.sum ] ; then
  go mod tidy
else
  go mod download
fi

CGO_ENABLED=0 go build -ldflags "$FLAGS"
echo "Version: $VERSION"
