ARG GO_VERSION
FROM golang:${GO_VERSION} as builder

ARG GOOS
ARG GOARCH
ARG GOARM

ENV GO111MODULE on
ENV GOOS ${GOOS}
ENV GOARCH ${GOARCH}
ENV GOARM ${GOARM}

# RClone
RUN git clone -b v1.66.0 https://github.com/rclone/rclone /go/src/github.com/rclone/rclone
WORKDIR /go/src/github.com/rclone/rclone
RUN go get ./...
RUN env ${BUILD_OPTS} go build

# Restic
RUN git clone -b v0.16.4 https://github.com/restic/restic /go/src/github.com/restic/restic
WORKDIR /go/src/github.com/restic/restic
RUN go get ./...
RUN go mod vendor
RUN GOOS= GOARCH= GOARM= go run -mod=vendor build.go || go run build.go

# Bivac
WORKDIR /go/src/github.com/camptocamp/bivac
COPY . .
RUN go mod vendor
RUN env ${BUILD_OPTS} make bivac

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y openssh-client procps && \
    apt-get dist-upgrade -y && \
	rm -rf /var/lib/apt/lists/*
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /go/src/github.com/camptocamp/bivac/bivac /bin/bivac
COPY --from=builder /go/src/github.com/camptocamp/bivac/providers-config.default.toml /
COPY --from=builder /go/src/github.com/restic/restic/restic /bin/restic
COPY --from=builder /go/src/github.com/rclone/rclone/rclone /bin/rclone
ENTRYPOINT ["/bin/bivac"]
CMD [""]
