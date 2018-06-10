IMAGE=tempusbreve/authsvc
SOURCE_ROOT=breve.us/authsvc

PUBLIC_SRC=../../../web/app/build/

ifeq ($(REVISION),)
	REVISION=$(shell git describe --dirty --first-parent --always --tags)
endif

all: help

help: ## Display help text.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

clean: ## Remove build artifacts
	-rm -rf ./authsvc ./public
	go clean ${SOURCE_ROOT}/...

prep: ## Prepare for docker build
	rsync -v -a --delete-after $(PUBLIC_SRC) ./public
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-s -w -X main.version=$(REVISION)" -o authsvc ${SOURCE_ROOT}/cmd/authsvc

image: prep ## Build Docker image
	docker build -t $(IMAGE):latest -t $(IMAGE):$(REVISION) .

push: clean image ## Push Docker images
	docker push $(IMAGE):$(REVISION)
	docker push $(IMAGE):latest

.PHONY: all help clean prep image push
