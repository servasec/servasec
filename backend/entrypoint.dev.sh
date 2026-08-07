#!/bin/sh
GIT_VERSION="${GIT_VERSION:-develop}"
LDFLAGS="-X github.com/servasec/servasec/backend/version.Version=${GIT_VERSION}"
if [ -n "$BUILD_TAGS" ]; then
  exec air --build.cmd "go build -tags '$BUILD_TAGS' -ldflags \"$LDFLAGS\" -o ./tmp/main"
else
  exec air --build.cmd "go build -ldflags \"$LDFLAGS\" -o ./tmp/main"
fi
