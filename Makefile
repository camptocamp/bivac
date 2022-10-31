DEPS = $(wildcard */*/*/*.go)
VERSION = $(shell git describe --always --dirty)
COMMIT_SHA1 = $(shell git rev-parse HEAD)
BUILD_DATE = $(shell date +%Y-%m-%d)

GO_VERSION = 1.19
RESTIC_VERSION = v0.14.0

all: lint vet test bivac

bivac: main.go $(DEPS)
	GO111MODULE=on CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) GOARM=$(GOARM) \
	  go build \
	    -a -ldflags="-s -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.commitSha1=$(COMMIT_SHA1)" \
	    -installsuffix cgo -o $@ $<
	@if [ "${GOOS}" = "linux" ] && [ "${GOARCH}" = "amd64" ]; then strip $@; fi

release: clean
	GO_VERSION=$(GO_VERSION) ./scripts/build-release.sh

docker-images: clean
	@if [ -z "$(IMAGE_NAME)" ]; then echo "IMAGE_NAME cannot be empty."; exit 1; fi
	export IMAGE_NAME=$(IMAGE_NAME)
	# Linux/amd64
	docker build --no-cache --pull -t $(IMAGE_NAME)-linux-amd64:$(IMAGE_VERSION) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GOOS=linux \
		--build-arg GOARCH=amd64 \
		--build-arg RESTIC_VERSION=$(RESTIC_VERSION)  \
		.
	docker push $(IMAGE_NAME)-linux-amd64:$(IMAGE_VERSION)
	# Linux/386
	docker build --no-cache --pull -t $(IMAGE_NAME)-linux-386:$(IMAGE_VERSION) \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg GOOS=linux \
		--build-arg GOARCH=386 \
		--build-arg RESTIC_VERSION=$(RESTIC_VERSION) \
		.
	docker push $(IMAGE_NAME)-linux-386:$(IMAGE_VERSION)
	# Linux/arm
	docker build --no-cache --pull -t $(IMAGE_NAME)-linux-arm:$(IMAGE_VERSION) \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg GOOS=linux \
		--build-arg GOARCH=arm \
		--build-arg GOARM=7 \
		--build-arg RESTIC_VERSION=$(RESTIC_VERSION) \
		.
	docker push $(IMAGE_NAME)-linux-arm:$(IMAGE_VERSION)
	# Linux/arm64
	docker build --no-cache --pull -t $(IMAGE_NAME)-linux-arm64:$(IMAGE_VERSION) \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg GOOS=linux \
		--build-arg GOARCH=arm64 \
		--build-arg GOARM=7 \
		--build-arg RESTIC_VERSION=$(RESTIC_VERSION) \
		.
	docker push $(IMAGE_NAME)-linux-arm64:$(IMAGE_VERSION)
	# Manifest
	docker manifest create $(IMAGE_NAME):$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-amd64:$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-386:$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-arm:$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-arm64:$(IMAGE_VERSION)
	docker manifest annotate $(IMAGE_NAME):$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-amd64:$(IMAGE_VERSION) --os linux --arch amd64
	docker manifest annotate $(IMAGE_NAME):$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-386:$(IMAGE_VERSION) --os linux --arch 386
	docker manifest annotate $(IMAGE_NAME):$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-arm:$(IMAGE_VERSION) --os linux --arch arm
	docker manifest annotate $(IMAGE_NAME):$(IMAGE_VERSION) \
		$(IMAGE_NAME)-linux-arm64:$(IMAGE_VERSION) --os linux --arch arm64
	docker manifest push $(IMAGE_NAME):$(IMAGE_VERSION)

lint:
	@GO111MODULE=off go get -u -v golang.org/x/lint/golint
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

.PHONY: all lint vet clean test
