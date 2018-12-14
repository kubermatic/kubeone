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

# Generate SSH key pair
if [ ! -f "$HOME/.ssh/id_rsa_kubeone_e2e" ]; then
 echo "Generating SSH key pair"
 ssh-keygen -f $HOME/.ssh/id_rsa_kubeone_e2e -N ''
 SSH_PUBLIC_KEY_FILE="$HOME/.ssh/id_rsa_kubeone_e2e.pub"
 SSH_PRIVATE_KEY_FILE="$HOME/.ssh/id_rsa_kubeone_e2e"
 export SSH_PUBLIC_KEY_FILE
 chmod 400 $SSH_PRIVATE_KEY_FILE
 eval `ssh-agent`
 ssh-add $SSH_PRIVATE_KEY_FILE
fi

# Build binaries
echo "Building kubeone ..."
make install


# Start the tests
echo "Runing E2E tests ..."
go test -race -tags=e2e -v -timeout 30m  ./test/e2e/...
