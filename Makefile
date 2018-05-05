

all: help

help:
	@echo "Help text goes here"

clean:
	go clean ./...

build:
	@.bin/build

format:
	@.bin/format

lint:
	@.bin/lint

verify: lint
	@.bin/verify

install-hooks:
	@.bin/install-pre-push-hook
