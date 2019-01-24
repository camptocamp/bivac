DEPS = $(wildcard */*/*/*.go)
VERSION = $(shell git describe --always --dirty)

all: bivac

bivac: main.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-s -X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

lint:
	@ go get -v github.com/golang/lint/golint
	@for file in $$(git ls-files '*.go' | grep -v '_workspace/'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vet: bivac.go
	go vet $<

imports: bivac.go
	dep ensure
	goimports -d $<

clean:
	rm -f bivac

test:
	richgo test -cover -coverprofile=coverage -v ./...

.PHONY: all imports lint vet clean test
