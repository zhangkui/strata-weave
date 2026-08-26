#!/usr/bin/env sh
set -eu
tag=${1:-strata-weave}
platform=${2:-linux/amd64}
docker build --platform "$platform" -f benzhi.Dockerfile -t "benzhi/$tag:latest" .
