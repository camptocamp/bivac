DEPS = $(wildcard */*/*/*.go)
VERSION = $(shell git describe --always --dirty)

all: lint vet test bivac

bivac: main.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-s -X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

lint:
	@go get -v golang.org/x/lint/golint
	@for file in $$(go list ./... | grep -v '_workspace/' | grep -v 'vendor'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vet: main.go
	go vet $<

imports: main.go
	dep ensure
	goimports -d $<

clean:
	rm -f bivac

test:
	go test -cover -coverprofile=coverage -v ./...

.PHONY: all imports lint vet clean test
