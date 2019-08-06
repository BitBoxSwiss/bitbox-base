#!/bin/bash
#
# Remove all build/ outputs.
#
set -eu

# Enable extended pattern matching operators like !$("file").
shopt -s extglob

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Remove all contents of build/, except README.md. 
rm -rf ${SCRIPT_DIR}/../build/!("README.md")
