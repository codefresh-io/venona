FROM node:10.13.0-alpine

WORKDIR /root/isser

RUN apk add --no-cache bash git openssh-client

COPY package.json ./

COPY yarn.lock ./

# install isser required binaries
RUN apk add --no-cache --virtual deps python make g++ krb5-dev && \
    yarn install --forzen-lockfile --production && \
    yarn cache clean && \
    apk del deps && \
    rm -rf /tmp/*

# copy app files
COPY . ./

# run application
CMD ["node", "index.js"]
