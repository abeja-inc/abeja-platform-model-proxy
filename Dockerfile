FROM ubuntu:18.04
RUN apt update && \
    DEBIAN_FRONTEND=noninteractive apt install -y wget curl && \
    DEBIAN_FRONTEND=noninteractive apt install --no-install-recommends -y \
    git make build-essential libssl-dev zlib1g-dev \
    libbz2-dev libreadline-dev libsqlite3-dev llvm \
    libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev && \
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz && \
    rm go1.21.5.linux-amd64.tar.gz && \
    /usr/local/go/bin/go install golang.org/x/tools/cmd/goimports@latest && \
    /usr/local/go/bin/go install github.com/rakyll/statik@latest

    # git clone https://github.com/abeja-inc/abeja-platform-model-proxy.git && \
    # cd /abeja-platform-model-proxy && \
    # /root/go/bin/statik -src=./runtime_src -p=runtime && \
    # /usr/local/go/bin/go build -o abeja-runner && \
    # cp abeja-runner /mnt/host
