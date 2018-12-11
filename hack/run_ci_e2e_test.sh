#!/usr/bin/env bash

# This script is mostly used in CI
# It installs dependencies and starts the tests

set -euxo pipefail

# Insatll dependencies
if ! [ -x "$(command -v terraform)" ]; then
(
  echo "Installing terraform"
  cd /tmp
  curl https://releases.hashicorp.com/terraform/0.11.10/terraform_0.11.10_linux_amd64.zip -Lo terraform.zip
  unzip -n terraform.zip terraform
  chmod +x terraform
  mv terraform /usr/local/bin
  rm terraform.zip
 )
fi

# Start the tests
echo "Runing E2E tests ..."
go test -race -tags=e2e -v -timeout 30m  ./test/e2e/...
