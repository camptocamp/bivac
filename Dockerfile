FROM golang:1.11 as builder
WORKDIR /go/src/github.com/camptocamp/bivac
COPY . .
RUN make bivac

FROM restic/restic:latest as restic

FROM busybox
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /go/src/github.com/camptocamp/bivac/bivac /bin/
COPY --from=builder /go/src/github.com/camptocamp/bivac/providers-config.default.toml /
COPY --from=restic /usr/bin/restic /bin/restic
ENTRYPOINT ["/bin/bivac"]
CMD [""]
