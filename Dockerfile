FROM golang:1.10 as builder
WORKDIR /go/src/github.com/camptocamp/bivac
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN make bivac

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/bivac/bivac /
ENTRYPOINT ["/bivac"]
CMD [""]
