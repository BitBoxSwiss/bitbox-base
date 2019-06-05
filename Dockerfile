FROM ubuntu:18.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    gcc \
    libc6-dev \
    make \
    curl \
    ca-certificates \
    git

RUN mkdir -p /opt/go_dist &&\
    curl -O https://dl.google.com/go/go1.12.5.linux-amd64.tar.gz &&\
    echo "aea86e3c73495f205929cfebba0d63f1382c8ac59be081b6351681415f4063cf go1.12.5.linux-amd64.tar.gz" | sha256sum -c &&\
    tar -xzf go1.12.5.linux-amd64.tar.gz -C /opt/go_dist

ENV GOPATH /opt/go
ENV GOROOT /opt/go_dist/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

ADD middleware/Makefile /tmp/
ADD middleware/scripts/go-get.sh /tmp/scripts/
RUN make -C  /tmp/ envinit

