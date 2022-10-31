#!/bin/bash

if [ -z "$GO_VERSION" ]; then
  GO_VERSION=latest
fi

docker pull "golang:$GO_VERSION"

VERSION=$(git describe --always --dirty)

PLATFORMS=(darwin linux windows)
ARCHITECTURES=(386 amd64)

mkdir -p release

for platform in "${PLATFORMS[@]}"; do
  for arch in "${ARCHITECTURES[@]}"; do
    rm -f bivac
    docker run -it --rm -w "/go/src/github.com/camptocamp/bivac" -v "$(pwd):/go/src/github.com/camptocamp/bivac" \
      -e "GOOS=${platform}" \
      -e "GOARCH=${arch}" \
      "golang:$GO_VERSION" make bivac
    sha256sum bivac >> release/SHA256SUM.txt
    zip "release/bivac_${VERSION}_${platform}_${arch}.zip" bivac
  done
done
