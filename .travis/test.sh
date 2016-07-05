#!/bin/bash


docker build -t testtag .

docker volume create --name test

docker run -v /var/run/docker.sock:/var/run/docker.sock:ro  --rm -ti \
  -e DUPLICITY_TARGET_URL="file:///root/.cache/duplicity/.test_backup" \
  testtag -l debug

if [ "$?" -ne 0 ]; then
  echo "E: Image test failed"
#  exit 1
fi
