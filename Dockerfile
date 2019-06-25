#
# The docker image produced by this config provides the build environment
# for BitBox Base.
#
FROM golang:1.12.5-stretch as bitbox-base
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    gcc \
    libc6-dev \
    make \
    ca-certificates 
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base
RUN mkdir build

# Build the middleware
FROM bitbox-base as middleware-builder
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base/middleware/
COPY middleware/scripts/ scripts/
RUN ./scripts/envinit.sh
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base
RUN rm -rf middleware
COPY scripts/. scripts/.
COPY middleware/. middleware/.
RUN make -C "middleware"

# Build the tools
FROM bitbox-base as middleware-tools
WORKDIR /go/src/github.com/digitalbitbox/bitbox-base
COPY scripts/. scripts/.
COPY tools/. tools/.
RUN make -C "tools"

# Final
FROM golang:1.12.5-stretch as final
ARG builder_uid
RUN adduser --disabled-password --gecos "" --uid ${builder_uid} builder
USER builder
COPY --from=middleware-builder /go/src/github.com/digitalbitbox/bitbox-base/build/. /opt/build/.
COPY --from=middleware-tools /go/src/github.com/digitalbitbox/bitbox-base/build/. /opt/build/.
