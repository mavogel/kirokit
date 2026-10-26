# The image scripts/linux-docker.sh runs commands in: what CI's Ubuntu jobs
# have for kirokit (Go, Node, Task, sops, WebKitGTK 4.1 for the desktop app,
# xvfb for its smoke start), plus NSIS and 7-Zip to build and list the Windows
# installer from Linux.
FROM node:22-bookworm AS node

FROM golang:1.27-bookworm
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
 && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config \
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
