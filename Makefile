VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo 0.1.0-dev)
export VERSION

.PHONY: sync-web run build build-all release tidy

sync-web:
	bash scripts/sync-web.sh

run:
	go run ./server -config-dir ./config

build: sync-web
	bash scripts/build.sh build

build-all:
	bash scripts/build.sh build-all

release:
	bash scripts/build.sh release

tidy:
	go mod tidy
