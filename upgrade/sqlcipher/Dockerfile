FROM ubuntu:20.04 AS build

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    git \
    libssl-dev \
    tclsh \
 && rm -rf /var/lib/apt/lists/*

ARG version=master

RUN cd tmp \
 && git clone --depth 1 --branch ${version} https://github.com/sqlcipher/sqlcipher.git \
 && cd sqlcipher \
 && ./configure --disable-tcl \
 && make sqlite3.c

FROM alpine

COPY --from=build /tmp/sqlcipher/sqlite3.* /tmp/sqlcipher/ext/userauth/*userauth.* /sqlcipher/
