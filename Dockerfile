FROM alpine:latest

RUN apk add --no-cache terraform

COPY dist/kubeone /usr/bin/kubeone

ENV SSH_AUTH_SOCK /run/ssh.sock
WORKDIR /mnt
