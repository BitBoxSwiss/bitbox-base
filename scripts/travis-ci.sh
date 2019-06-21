#!/usr/bin/env bash
#
# Run CI scripts to build / test project.
#
set -euo pipefail

TRAVIS_BUILD_DIR=${TRAVIS_BUILD_DIR:-"$(pwd)"}
# TODO(hkjn): We could 'make build-all' here if we can resolve
# remaining issues with building Armbian images in a dockerized
# workflow on Travis:
# https://github.com/digitalbitbox/bitbox-base/issues/39#issuecomment-501343881
make docker-build-go
make python-style-check
