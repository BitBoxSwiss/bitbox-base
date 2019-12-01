#
# The docker image produced by this config provides the build environment
# for BitBoxBase.
#
FROM golang:1.13.4-stretch as bitbox-base
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    gcc \
    libc6-dev \
    make \
    ca-certificates
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base
RUN mkdir -p bin/go/

# Build the middleware
FROM bitbox-base as middleware-builder
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base/middleware/
COPY middleware/contrib/ contrib/
RUN ./contrib/envinit.sh
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base
RUN rm -rf middleware
COPY contrib/. contrib/.
COPY middleware/. middleware/.
RUN make -C "middleware"

# Build the tools
FROM bitbox-base as middleware-tools
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base
COPY contrib/. contrib/.
COPY tools/. tools/.
RUN make -C "tools"

# Final
FROM golang:1.13.4-stretch as final

COPY --from=middleware-builder /go/src/github.com/digitalbitbox/bitbox-base/bin/go/. /opt/build/.
COPY --from=middleware-tools /go/src/github.com/digitalbitbox/bitbox-base/bin/go/. /opt/build/.
