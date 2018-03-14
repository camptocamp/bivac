FROM golang:1.10 as builder
WORKDIR /go/src/github.com/camptocamp/conplicity
COPY . .
# TODO: use vendoring
RUN go get golang.org/x/tools/cmd/goimports \
           github.com/Sirupsen/logrus \
	       github.com/docker/docker/api \
	       github.com/go-ini/ini \
	       github.com/jessevdk/go-flags \
		   golang.org/x/net/context
RUN make conplicity

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/conplicity/conplicity /
ENTRYPOINT ["/conplicity"]
CMD [""]
