#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

FROM golang:1

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive
ENV GOPRIVATE=github.com/sabio-engineering-product/*
# Configure apt, install packages and tools
RUN apt-get update \
  && apt-get -y install --no-install-recommends apt-utils 2>&1 \
  #
  # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
  && apt-get -y install git graphviz gv unzip procps lsb-release \
  #
  # Install gocode-gomod
  && go get -x -d github.com/stamblerre/gocode 2>&1 \
  && go build -o gocode-gomod github.com/stamblerre/gocode \
  && mv gocode-gomod $GOPATH/bin/ \
  #
  # Install Go tools
  && go get -u -v \
  github.com/mdempsky/gocode \
  github.com/uudashr/gopkgs/cmd/gopkgs \
  github.com/ramya-rao-a/go-outline \
  github.com/acroca/go-symbols \
  github.com/godoctor/godoctor \
  golang.org/x/tools/cmd/guru \
  golang.org/x/tools/cmd/gorename \
  github.com/rogpeppe/godef \
  github.com/zmb3/gogetdoc \
  github.com/haya14busa/goplay/cmd/goplay \
  github.com/sqs/goreturns \
  github.com/josharian/impl \
  github.com/davidrjenni/reftools/cmd/fillstruct \
  github.com/fatih/gomodifytags \
  github.com/cweill/gotests/... \
  golang.org/x/tools/cmd/goimports \
  golang.org/x/lint/golint \
  golang.org/x/tools/gopls \
  github.com/alecthomas/gometalinter \
  honnef.co/go/tools/... \
  github.com/mgechev/revive \
  github.com/derekparker/delve/cmd/dlv 2>&1 \
  #
  # Clean up
  && apt-get autoremove -y \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /bin/docker-entrypoint.sh
RUN chmod +x /bin/docker-entrypoint.sh
RUN curl https://releases.hashicorp.com/consul/1.6.2/consul_1.6.2_linux_amd64.zip --output /tmp/consul.zip \
  && unzip /tmp/consul.zip -d /tmp/ \
  && mv /tmp/consul /usr/bin/. \
  && rm -f /tmp/consul.zip
