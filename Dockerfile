FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl \
    git \
    ca-certificates \
    iproute2 \
    iputils-ping && \
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y nodejs 

RUN npm install -g openclaw && \
    npm cache clean --force && \
    rm -rf /root/.npm /var/lib/apt/lists/* /tmp/*

COPY guest /usr/local/bin/guest
RUN chmod +x /usr/local/bin/guest
