FROM golang:1.13.4-alpine

RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    git \
    make \
    unzip \
    upx

ENV TERRAFORM_VERSION "0.12.13"
RUN curl -fL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip | \
    funzip > /usr/local/bin/terraform
RUN chmod +x /usr/local/bin/terraform

ENV KUBETEST_COMMIT "eb1a87c9083a31f9a9dda7822cb674f9f8d99439"
ENV TEST_ROOT_DIR "/opt/kube-test"
ENV GO111MODULE=on

# Install the kubetest binary
RUN go get -v k8s.io/test-infra/kubetest@${KUBETEST_COMMIT} && \
    go clean -modcache -cache

COPY install-kube-tests.sh /opt/install-kube-tests.sh
RUN /opt/install-kube-tests.sh
