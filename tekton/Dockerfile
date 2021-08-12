# syntax=docker/dockerfile:1.3
FROM alpine:3 as download

RUN set -ex \
 && mkdir -p /opt/bin \
 && apk add --no-cache curl \
 && cd /tmp \
 && curl -fsSL https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 -o jq \
 && chmod +x jq \
 && mv jq /opt/bin \
 && curl -fsSL https://github.com/cli/cli/releases/download/v1.14.0/gh_1.14.0_linux_amd64.tar.gz -o gh.tar.gz \
 && tar xvzf gh.tar.gz \
 && mv gh_*_linux_amd64/bin/gh /opt/bin/gh \
 && rm -rf *

FROM docker.io/golang:1.16.7

ENV GOBIN=/opt/bin
ENV PATH=/opt/bin:$PATH

COPY Makefile /tools/

RUN --mount=type=cache,target=/go set -ex \
 && mkdir -p /opt/bin \
 && cd /tools \
 && make kustomize controller-gen envtest \
 && go install github.com/golang-jwt/jwt/v4/cmd/jwt@latest \
 && mv bin/* /opt/bin/

COPY --from=download /opt/bin/* /opt/bin/
