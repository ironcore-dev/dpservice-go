#!/usr/bin/env bash

NET_DPSERVICE_REVISION="main"

set -e

BASEDIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

NET_DPSERVICE_DIR="$(mktemp -d)"
trap 'rm -rf $NET_DPSERVICE_DIR' EXIT

function clone() {
  cd "$NET_DPSERVICE_DIR"
  git init
  git remote add origin git@github.com:onmetal/net-dpservice.git
  git fetch origin "$NET_DPSERVICE_REVISION" --depth=1
  git reset --hard FETCH_HEAD
}

echo "Cloning repository"
if ! err="$(clone 2>&1)"; then
  echo "Error cloning repository:"
  echo "$err"
fi

echo "Generating protobuf"
protoc --proto_path="$NET_DPSERVICE_DIR" \
  --go_out="$BASEDIR"/.. \
  --go_opt=paths=source_relative \
  --go-grpc_out="$BASEDIR"/.. \
  --go-grpc_opt=paths=source_relative \
  "$NET_DPSERVICE_DIR"/proto/dpdk.proto

for file in "$BASEDIR"/../proto/*.pb.go; do
  boilerplate="$(cat "$BASEDIR"/boilerplate.go.txt)"
  echo -e "$boilerplate\n$(cat "$file")" > "$file"
done
