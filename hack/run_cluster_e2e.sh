#!/usr/bin/env bash

# Source common.sh
source $(dirname "${BASH_SOURCE}")/common.sh

create_kubeconfig
start_tests
