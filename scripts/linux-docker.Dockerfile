# The image scripts/linux-docker.sh runs commands in: what CI's Ubuntu jobs
# have for kirokit (Go, Node, Task, sops, GTK 4 / WebKitGTK 6.0 for the desktop
# app, xvfb for its smoke start), plus NSIS and 7-Zip to build and list the
# Windows installer from Linux.
#
# Ubuntu 24.04 is the oldest system the Linux app supports (Wails v3 needs
# GTK 4.14) and the release's Linux build host, so local builds match CI's.
FROM node:26-bookworm AS node
FROM golang:1.27-bookworm AS go

FROM ubuntu:24.04
ENV GOPATH=/go PATH=/go/bin:/usr/local/go/bin:$PATH
COPY --from=go /usr/local/go /usr/local/go
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
 && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates curl git gcc libc6-dev \
      libgtk-4-dev libwebkitgtk-6.0-dev pkg-config \
      xvfb xauth nsis p7zip-full \
 && rm -rf /var/lib/apt/lists/*

ARG SOPS_VERSION=3.13.3
RUN arch="$(dpkg --print-architecture)" \
 && cd /tmp \
 && curl -fsSLO "https://github.com/getsops/sops/releases/download/v${SOPS_VERSION}/sops-v${SOPS_VERSION}.linux.${arch}" \
 && curl -fsSLO "https://github.com/getsops/sops/releases/download/v${SOPS_VERSION}/sops-v${SOPS_VERSION}.checksums.txt" \
 && grep " sops-v${SOPS_VERSION}.linux.${arch}\$" "sops-v${SOPS_VERSION}.checksums.txt" | sha256sum -c - \
 && install -m 0755 "sops-v${SOPS_VERSION}.linux.${arch}" /usr/local/bin/sops \
 && rm -f /tmp/sops-*

RUN go install github.com/go-task/task/v3/cmd/task@v3.54.0 \
 && git config --system --add safe.directory '*'
