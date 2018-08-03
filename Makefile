# VERSION_TAG:=$(shell git describe --abbrev=0 --tags || echo "0.1")
VERSION_TAG:=$(echo "0.1")
LDFLAGS:=-ldflags "-X github.com/kmacoskey/taos/app.Version=${VERSION_TAG}"

.PHONY: test clean

default: build

test: clean
	ginkgo -r -failFast
