FROM golang:1.12 as builder
WORKDIR /go/src/github.com/camptocamp/bivac
COPY . .
RUN make bivac

FROM restic/restic:latest as restic

FROM alpine:latest as rclone
RUN wget https://downloads.rclone.org/rclone-current-linux-amd64.zip
RUN unzip rclone-current-linux-amd64.zip
RUN cp rclone-*-linux-amd64/rclone /usr/bin/
RUN chown root:root /usr/bin/rclone
RUN chmod 755 /usr/bin/rclone

FROM debian
RUN apt-get update && \
    apt-get install -y openssh-client procps && \
	rm -rf /var/lib/apt/lists/*
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /go/src/github.com/camptocamp/bivac/bivac /bin/
COPY --from=builder /go/src/github.com/camptocamp/bivac/providers-config.default.toml /
COPY --from=restic /usr/bin/restic /bin/restic
COPY --from=rclone /usr/bin/rclone /bin/rclone
ENTRYPOINT ["/bin/bivac"]
CMD [""]
