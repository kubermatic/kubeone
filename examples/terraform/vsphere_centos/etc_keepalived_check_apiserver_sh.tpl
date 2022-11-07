#!/bin/sh

errorExit() {
    echo "*** $*" 1>&2
    exit 1
}

curl --silent --max-time 2 --insecure https://localhost:6443/healthz -o /dev/null || errorExit "Error GET https://localhost:6443/healthz"
if ip addr | grep -q ${APISERVER_VIP}; then
    curl --silent --max-time 2 --insecure https://${APISERVER_VIP}:6443/healthz -o /dev/null || errorExit "Error GET https://${APISERVER_VIP}:6443/healthz"
fi
