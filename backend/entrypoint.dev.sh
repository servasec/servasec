#!/bin/sh
if [ -n "$BUILD_TAGS" ]; then
  exec air --build.cmd "go build -tags '$BUILD_TAGS' -o ./tmp/main"
else
  exec air
fi
