#!/usr/bin/env bash

set -x

mkdir /tmp/gobetween
cd /tmp/gobetween
curl -L -o gobetween_0.7.0_linux_amd64.tar.gz \
    https://github.com/yyyar/gobetween/releases/download/0.7.0/gobetween_0.7.0_linux_amd64.tar.gz
tar xvf gobetween_0.7.0_linux_amd64.tar.gz
mv gobetween /usr/local/sbin/gobetween
chown root:root /usr/local/sbin/gobetween

cat <<EOF > /etc/systemd/system/gobetween.service
[Unit]
Description=Gobetween - modern LB for cloud era
Documentation=https://github.com/yyyar/gobetween/wiki
After=network.target remote-fs.target nss-lookup.target

[Service]
Type=simple
PIDFile=/run/gobetween.pid
ExecStart=/usr/local/sbin/gobetween -c /etc/gobetween.toml
ExecReload=/bin/kill -s HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
PrivateTmp=true
User=nobody
Group=nogroup

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable gobetween.service
systemctl start gobetween.service
