#!/bin/bash
set -eo pipefail

mkdir -p ./bin
GO111MODULE=on \
CGO_ENABLED=0 \
GOOS=linux \
go \
    build -v -a \
    -o ./bin/replay-spreadsheet ./cmd/replay-spreadsheet

TAG=latest
IMAGE=paralin/replay-spreadsheet
IMAGE_TAG=${IMAGE}:${TAG}
IMAGE_TAG_ALT=paralin/replay-spreadsheet:$TAG

docker build -f Dockerfile.bin -t "${IMAGE_TAG_ALT}" ./
gcloud docker -- push ${IMAGE_TAG_ALT}
