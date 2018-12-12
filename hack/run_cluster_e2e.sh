#!/usr/bin/env bash

# NOTE:
#
# This file is depreciated and will be replaced by run_ci_e2e_test.sh

# Source common.sh
source $(dirname "${BASH_SOURCE}")/common.sh

create_kubeconfig
start_tests
