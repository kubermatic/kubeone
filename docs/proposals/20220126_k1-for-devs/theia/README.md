
This document show how to get Theia up and running on a new ubuntu 20.04 machine.

# get build environment
```bash

# install nodejs
curl -sL https://deb.nodesource.com/setup_12.x | sudo -E bash -
DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs

# install yarn
curl -sL https://dl.yarnpkg.com/debian/pubkey.gpg | gpg --dearmor | tee /usr/share/keyrings/yarnkey.gpg >/dev/null
echo "deb [signed-by=/usr/share/keyrings/yarnkey.gpg] https://dl.yarnpkg.com/debian stable main" | tee /etc/apt/sources.list.d/yarn.list
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y yarn build-essential libsecret-1-0
```

# build
```bash
yarn
```

# settings file
```bash
mkdir /root/.theia
cp settings.json /root/.theia/settings.json
```

# run
```bash
yarn start <WORKSPACE-DIR> --hostname 0.0.0.0 --port 8080
yarn start /home/hubert/Desktop/theia/workspace --hostname 0.0.0.0 --port 8080
```