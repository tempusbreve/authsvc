IMAGE=tempusbreve/authsvc
SOURCE_ROOT=breve.us/authsvc

ifeq ($(REVISION),)
	DIRTY=$(shell git diff-index --quiet HEAD || echo "-dirty")
	REVISION=$(shell git rev-parse --short HEAD)$(DIRTY)
endif

all: image

clean:
	-rm ./api
	go clean ${SOURCE_ROOT}/authsvc/...

image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-s -w" -o api ${SOURCE_ROOT}/cmd/api
	docker build -t $(IMAGE):latest -t $(IMAGE):$(REVISION) .

push: clean image
	docker push $(IMAGE):$(REVISION)
	docker push $(IMAGE):latest

