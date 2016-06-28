DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always --dirty)

all: test conplicity conplicity.1

conplicity: conplicity.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

conplicity.1: conplicity
	./conplicity -m > $@

lint:
	@ go get -v github.com/golang/lint/golint
	@for file in $$(git ls-files '*.go' | grep -v '_workspace/'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vet: conplicity.go
	go vet $<

imports: conplicity.go
	goimports -d $<

test: lint vet imports
	go test ./...

coverage:
	rm -rf coverage/
	mkdir -p coverage/
	go test -coverprofile=coverage.out
	for i in handler providers util volume; do \
		go test -coverprofile=coverage/$$i.coverage.out ./$$i/; \
		tail -n +2 coverage/$$i.coverage.out >> coverage.out; \
  done

clean:
	rm -f conplicity conplicity.1
