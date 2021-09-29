#!/usr/bin/env bash

set -o monitor
set -exuo pipefail

function cleanup {
  set +e
  echo "starting sleep"
  sleep 300
  echo "sleep done"
}
trap cleanup EXIT

sleep 1000

