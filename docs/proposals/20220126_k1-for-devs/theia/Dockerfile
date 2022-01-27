# build
FROM node:12.18.3 as build
RUN apt-get -qq update && apt-get install -y libsecret-1-dev
WORKDIR /theia
COPY package.json ./package.json
RUN yarn --pure-lockfile && \
    NODE_OPTIONS="--max_old_space_size=4096" yarn theia build && \
    yarn theia download:plugins && \
    yarn --production && \
    yarn autoclean --init && \
    echo *.ts >> .yarnclean && \
    echo *.ts.map >> .yarnclean && \
    echo *.spec.* >> .yarnclean && \
    yarn autoclean --force && \
    yarn cache clean

# run    
FROM node:12.18.3
WORKDIR /theia

# install kubectl
RUN apt-get -qq update && apt-get install -y libsecret-1-0 curl bash-completion
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# theia
COPY --from=build /theia /theia
ENV SHELL=/bin/bash \
    THEIA_DEFAULT_PLUGINS=local-dir:/theia/plugins

# tooling
RUN apt-get install bash-completion -y
COPY .bashrc /root/.bashrc    
COPY settings.json /root/.theia/settings.json

ENTRYPOINT [ "node", "/theia/src-gen/backend/main.js", "<WORKSPACE-DIR>", "--hostname=0.0.0.0" ]