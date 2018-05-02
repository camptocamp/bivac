FROM golang:1.10 as builder
WORKDIR /go/src/github.com/camptocamp/conplicity
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN make conplicity

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/conplicity/conplicity /
ENTRYPOINT ["/conplicity"]
CMD [""]
