#!/bin/sh

TAG=`git describe --tags --always --dirty | sed -e 's/^v//'`
IMAGE=gcr.io/astute-synapse-332322/core-ledger-service
PLATFORMS=linux/amd64,linux/arm64

echo "Build tag $TAG"

docker buildx build --platform=$PLATFORMS -t $IMAGE -t $IMAGE:$TAG --output type=image,push=true  . 
docker buildx build --platform=$PLATFORMS -t $IMAGE:$TAG-alpine --output type=image,push=true  . 
