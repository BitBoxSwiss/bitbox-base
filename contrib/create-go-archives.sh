#!/bin/bash
# 
# create archives for go binaries
# 
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${SCRIPT_DIR}/../bin/go/"

set -x
tar czvf bbbconfgen.tar.gz bbbconfgen
tar czvf bbbfancontrol.tar.gz bbbfancontrol bbbfancontrol.service
tar czvf bbbmiddleware.tar.gz bbbmiddleware
tar czvf bbbsupervisor.tar.gz bbbsupervisor bbbsupervisor.service

set +x
ls -lh ./*.tar.gz

cd "${SCRIPT_DIR}"
