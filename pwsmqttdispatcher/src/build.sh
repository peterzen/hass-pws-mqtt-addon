#!/bin/bash

case "$BUILD_ARCH" in \
    "amd64")    GOARCH=amd64 ;; \
    "aarch64")  GOARCH=arm64 ;; \
    "armhf")    GOARCH=arm ;; \
    "armv7")    GOARCH=arm ;; \
    "i386")     GOARCH=386 
esac

CGO_ENABLED=0 GO111MODULE=on GOARCH=$GOARCH GOOS=linux go build


