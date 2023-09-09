FROM golang:1.21 as go

FROM cloudfoundry/cflinuxfs3

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

ARG BOOST_VERSION=1_59_0
ARG BOOST_DIGEST=47f11c8844e579d02691a607fbd32540104a9ac7a2534a8ddaef50daf502baac
RUN curl -Ls "https://jenkins.percona.com/downloads/boost/boost_${BOOST_VERSION}.tar.gz" >/boost.tgz \
  && echo "${BOOST_DIGEST}  /boost.tgz" \
  | sha256sum -c - \
  && tar -C /usr/local/src -xf /boost.tgz \
  && rm -f /boost.tgz

ARG PERCONA_VERSION=5.7.33-36
ARG PERCONA_DIGEST=964e32f4a1e235421e26be81d2d24f9e659d8ef3cbd9ae6c3e85fe545bedfd5b
RUN curl -Ls "https://www.percona.com/downloads/Percona-Server-${PERCONA_VERSION%.*-*}/Percona-Server-${PERCONA_VERSION}/source/tarball/percona-server-${PERCONA_VERSION}.tar.gz" >/percona.tgz \
    && echo "${PERCONA_DIGEST} /percona.tgz" \
    | sha256sum -c - \
    && tar -C /usr/local/src -xf /percona.tgz \
    && rm -f /percona.tgz
