#!/usr/bin/env bash

set -e

dir=$(cd `dirname $0` && cd .. && cd .. && cd .. && cd .. && pwd)
export GOPATH=${dir}
export GOARCH=amd64
export GOOS=darwin
export BLT_HOME=$HOME/.blt

if [[ -z "${1}" ]]; then
  echo "USAGE: $0 <version>"
  exit 1
fi

version=$1

rm -rf out
mkdir -p out

go build \
  -ldflags "-X github.com/aemengo/blt/cmd.version=${version}" \
  -o ./out/blt-${version}-${GOOS}-${GOARCH} \
  github.com/aemengo/blt

echo "${version}" > ${BLT_HOME}/assets/version
tar czf ./out/bosh-lit-assets.tgz -C ${BLT_HOME} assets

sha=$(shasum -a 1 ./out/bosh-lit-assets.tgz | awk '{print $1}')
echo ${sha} > ./out/bosh-lit-assets.tgz.sha1

sha256=$(shasum -a 256 ./out/blt-*-darwin-amd64 | awk '{print $1}')
cat <<EOF
class Blt < Formula
  desc "CLI for managing a local BOSH environment."
  homepage "https://github.com/aemengo/blt"
  version "${version}"
  url "https://github.com/aemengo/blt/releases/download/#{version}/blt-#{version}-darwin-amd64"
  sha256 "${sha256}"

  depends_on arch: :x86_64
  depends_on 'linuxkit/linuxkit/linuxkit'
  depends_on 'cloudfoundry/tap/bosh-cli'

  def install
    # system "/usr/local/bin/brew", "cask", "install", "docker"

    binary_name = "blt"
    bin.install "blt-#{version}-darwin-amd64" => binary_name
  end

  test do
    system "#{bin}/#{binary_name} --help"
  end
end

EOF