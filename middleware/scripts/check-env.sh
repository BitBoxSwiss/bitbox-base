#!/usr/bin/env bash

set -euo pipefail

GOPATH=${GOPATH:-""}
REPO_ROOT=${1:-"bug-missing-args"}

BAD_OR_MISSING_GOPATH_MSG="
================================================================================
fatal: ${0}:
  \${GOPATH} needs to be set to build Go programs, and the repository
  needs to checked out inside \${GOPATH}/src.

  Your current \${GOPATH} is:
  \"${GOPATH}\"

  Your repository is checked out at:
  \"${REPO_ROOT}\"

  For more information, see:
  https://github.com/golang/go/wiki/GOPATH#vendor-directories-ignored-outside-of-gopath
================================================================================
"

function log_error_and_exit() {
  echo "${BAD_OR_MISSING_GOPATH_MSG}" >&2
  exit 1
}


function check_gopath() {
  if [[ ! "${GOPATH}" ]]; then
    log_error_and_exit
  fi

  # TODO(hkjn): We should probably switch to Go modules to get away from requiring that
  # this repo is check out under ${GOPATH}/src, if Go 1.11+ can be assumed for the
  # build environment:
  # https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away

  # ${REPO_ROOT} needs to be under ${GOPATH}/src
  if ! [[ "${REPO_ROOT}" =~ "${GOPATH}/src" ]]; then
    log_error_and_exit
  fi
}

check_gopath
# environment seems fine!
exit 0

