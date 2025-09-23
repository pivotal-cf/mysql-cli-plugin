FROM golang:1.25 as go

FROM cloudfoundry/cflinuxfs4

COPY --from=go /usr/local/go /usr/local/go

ENV PATH=/usr/local/go/bin:$PATH GOPATH=/go GOCACHE=/go

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install --assume-yes --no-install-recommends \
    cmake \
    curl \
    g++ \
    make \
    libncurses5-dev \
    libssl-dev \
    zlib1g-dev \
  && rm -fr /var/lib/apt/lists/*

ARG PERCONA_VERSION=8.4.5-5
ARG PERCONA_DIGEST=8b47ff35dc2a6e7eaacaa2d204ae456c15b5d9953360ccb6250da8d68d98f6af
# //downloads.percona.com/downloads/Percona-Server-8.0/Percona-Server-8.0.43-34/source/tarball/percona-server-8.0.43-34.tar.gz
RUN curl -LSs "https://downloads.percona.com/downloads/Percona-Server-${PERCONA_VERSION%.*-*}/Percona-Server-${PERCONA_VERSION}/source/tarball/percona-server-${PERCONA_VERSION}.tar.gz" > /percona.tgz \
    && echo "${PERCONA_DIGEST} /percona.tgz" \
    | sha256sum -c - \
    && tar -C /usr/local/src -xf /percona.tgz \
    && rm -f /percona.tgz
