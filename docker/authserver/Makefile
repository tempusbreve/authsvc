IMAGE=tempusbreve/authserver
SOURCE_ROOT=breve.us/authsvc

ifeq ($(REVISION),)
	REVISION=$(shell git describe --dirty --first-parent --always --tags)
endif

all: image

clean:
	-rm ./authserver
	go clean ${SOURCE_ROOT}/...

image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-s -w -X main.version=$(REVISION)" -o authserver ${SOURCE_ROOT}/cmd/authserver
	docker build -t $(IMAGE):latest -t $(IMAGE):$(REVISION) .

push: clean image
	docker push $(IMAGE):$(REVISION)
	docker push $(IMAGE):latest

