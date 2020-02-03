#!/bin/bash

VERSION=$(git describe --always --dirty)

TARGET_IMAGE_REPO=$(echo $IMAGE_NAME | cut -d ":" -f 1)

TARGET_PLATFORMS=(
  ${TARGET_IMAGE_REPO}-linux-amd64:${VERSION}
  ${TARGET_IMAGE_REPO}-linux-arm:${VERSION}
  ${TARGET_IMAGE_REPO}-linux-arm64:${VERSION}
)


for target_platform in ${TARGET_PLATFORMS[@]}; do
  PLATFORM=$(printf '%s\n' "${target_platform//$TARGET_IMAGE_REPO/}" | cut -d ":" -f 1 | cut -c2-)
  GOOS=$(echo $PLATFORM | cut -d "-" -f 1)
  GOARCH=$(echo $PLATFORM | cut -d "-" -f 2)

  docker build --no-cache --pull -t ${target_platform} \
    --build-arg GO_VERSION=$GO_VERSION \
    --build-arg BUILD_OPTS="GOOS=${GOOS} GOARCH=${GOARCH}" .
  docker push ${target_platform}
  unset PLATFORM GOOS GOARCH
done

docker manifest create ${IMAGE_NAME} ${TARGET_PLATFORMS[@]}
for target_platform in ${TARGET_PLATFORMS[@]}; do
  PLATFORM=$(printf '%s\n' "${target_platform//$TARGET_IMAGE_REPO/}" | cut -d ":" -f 1 | cut -c2-)
  GOOS=$(echo $PLATFORM | cut -d "-" -f 1)
  GOARCH=$(echo $PLATFORM | cut -d "-" -f 2)

  docker manifest annotate ${IMAGE_NAME} ${target_platform} --os ${GOOS} --arch ${GOARCH}
  unset PLATFORM GOOS GOARCH
done

docker manifest push --purge ${IMAGE_NAME}
docker image rm ${TARGET_PLATFORMS[@]}
