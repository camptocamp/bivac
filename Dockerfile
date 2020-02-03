ARG GO_VERSION
FROM golang:${GO_VERSION} as builder

ARG BUILD_OPTS

# RClone
RUN go get github.com/rclone/rclone
WORKDIR /go/src/github.com/rclone/rclone
RUN env ${BUILD_OPTS} go build

# Restic
RUN go get github.com/restic/restic
WORKDIR /go/src/github.com/restic/restic
RUN env ${BUILD_OPTS} make restic

# Bivac
WORKDIR /go/src/github.com/camptocamp/bivac
COPY . .
RUN env ${BUILD_OPTS} make bivac

FROM debian
RUN apt-get update && \
    apt-get install -y openssh-client procps && \
	rm -rf /var/lib/apt/lists/*
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /go/src/github.com/camptocamp/bivac/bivac /bin/bivac
COPY --from=builder /go/src/github.com/camptocamp/bivac/providers-config.default.toml /
COPY --from=builder /go/src/github.com/restic/restic /bin/restic
COPY --from=builder /go/src/github.com/rclone/rclone /bin/rclone
ENTRYPOINT ["/bin/bivac"]
CMD [""]
