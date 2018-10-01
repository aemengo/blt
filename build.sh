#!/usr/bin/env bash

set -e

dir=$(cd `dirname $0` && cd .. && cd .. && cd .. && cd .. && pwd)
export GOPATH=${dir}
export GOARCH=amd64
export GOOS=darwin
export BLT_HOME=${HOME:-}/.blt

if [[ -z "${1}" ]]; then
  echo "USAGE: $0 <version>"
  exit 1
fi

version=$1

mkdir -p out

go build \
  -ldflags "-X github.com/aemengo/blt/cmd.version=${version}" \
  -o ./out/blt-${version}-${GOOS}-${GOARCH} \
  github.com/aemengo/blt

echo "${version}" > ${BLT_HOME}/assets/version
tar czf ./out/bosh-lit-assets.tgz -C ${BLT_HOME} assets

sha=$(shasum -a 1 ./out/bosh-lit-assets.tgz | awk '{print $1}')
echo ${sha} > ./out/bosh-lit-assets.tgz.sha1
