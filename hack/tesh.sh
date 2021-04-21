#!/usr/bin/env bash

trap "rm server;kill 0" EXIT
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
echo $ROOT
go build -o "${ROOT}"/server "${ROOT}"
"${ROOT}"/server -port=8001 -hostname=0.0.0.0 &
"${ROOT}"/server -port=8002 -hostname=0.0.0.0 &
"${ROOT}"/server -port=8003 -hostname=0.0.0.0 &
"${ROOT}"/server -port=8080 -hostname=0.0.0.0 -proxy=true &

wait
