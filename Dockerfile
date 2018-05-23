FROM golang:1.10 as builder
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/camptocamp/bivac
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only
COPY . .
RUN make bivac

FROM scratch
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /go/src/github.com/camptocamp/bivac/bivac /
ENTRYPOINT ["/bivac"]
CMD [""]
