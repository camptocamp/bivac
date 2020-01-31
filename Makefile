DEPS = $(wildcard */*/*/*.go)
VERSION = $(shell git describe --always --dirty)
COMMIT_SHA1 = $(shell git rev-parse HEAD)
BUILD_DATE = $(shell date +%Y-%m-%d)

all: lint vet test bivac

bivac: main.go $(DEPS)
	GO111MODULE=on CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) GOARM=$(GOARM) \
	  go build -mod=vendor -a \
		  -ldflags="-s -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.commitSha1=$(COMMIT_SHA1)" \
	    -installsuffix cgo -o $@ $<
	@if [ "${GOOS}" = "linux" ] && [ "${GOARCH}" = "amd64" ]; then strip $@; fi

release: clean
	GO_VERSION=1.12 ./scripts/build-release.sh

docker-images: clean
	@if [ -z "$(IMAGE_NAME)" ]; then echo "IMAGE_NAME cannot be empty."; exit 1; fi
	GO_VERSION=1.12 IMAGE_NAME=$(IMAGE_NAME) ./scripts/build-docker-images.sh

lint:
	@go get -u -v golang.org/x/lint/golint
	@for file in $$(go list ./... | grep -v '_workspace/' | grep -v 'vendor'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vet: main.go
	go vet $<

clean:
	git clean -fXd -e \!vendor -e \!vendor/**/* && rm -f ./bivac

test:
	go test -cover -coverprofile=coverage -v ./...

vendor:
	go mod vendor

.PHONY: all vendor lint vet clean test
