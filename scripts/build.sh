#!/usr/bin/env sh
set -eu

APP=${APP:-zjsh}
BIN_DIR=${BIN_DIR:-bin}
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || printf dev)}
LDFLAGS="-s -w -X github.com/saweima12/zjsh/internal/version.Version=${VERSION}"

mkdir -p "${BIN_DIR}"
CGO_ENABLED=0 go build \
  -ldflags "${LDFLAGS}" \
  -o "${BIN_DIR}/${APP}" \
  ./cmd/zjsh

printf 'built %s (%s)\n' "${BIN_DIR}/${APP}" "${VERSION}"
