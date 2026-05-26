#!/usr/bin/env sh
set -eu

APP=${APP:-zjsh}
BIN_DIR=${BIN_DIR:-bin}
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || printf dev)}
GOOS_VALUE=${GOOS:-$(go env GOOS)}
GOARCH_VALUE=${GOARCH:-$(go env GOARCH)}

EXT=""
if [ "${GOOS_VALUE}" = "windows" ]; then
  EXT=".exe"
fi

OUTPUT=${OUTPUT:-"${BIN_DIR}/${APP}${EXT}"}

LDFLAGS="-s -w -X github.com/tassis/zjsh/internal/version.Version=${VERSION}"

mkdir -p "$(dirname "${OUTPUT}")"

CGO_ENABLED=0 GOOS="${GOOS_VALUE}" GOARCH="${GOARCH_VALUE}" go build \
  -trimpath \
  -ldflags "${LDFLAGS}" \
  -o "${OUTPUT}" \
  ./cmd/zjsh

printf 'built %s (%s, %s/%s)\n' "${OUTPUT}" "${VERSION}" "${GOOS_VALUE}" "${GOARCH_VALUE}"
