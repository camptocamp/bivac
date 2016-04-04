FROM golang:alpine
RUN apk update && \
    apk add git && \
    go get github.com/camptocamp/conplicity && \
    apk del git
WORKDIR /go/bin
ENTRYPOINT ["conplicity"]
