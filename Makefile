DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always --dirty)

all: test bivac bivac.1

bivac: bivac.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

#bivac.1: bivac
#./bivac -m > $@

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

test: imports lint vet
	go test -v ./...

coverage:
	rm -rf *.out
	go test -coverprofile=coverage.out
	for i in config handler; do \
	 	go test -coverprofile=$$i.coverage.out github.com/camptocamp/bivac/$$i; \
		tail -n +2 $$i.coverage.out >> coverage.out; \
		done

clean:
	rm -f bivac bivac.1

.PHONY: all imports lint vet test coverage clean
