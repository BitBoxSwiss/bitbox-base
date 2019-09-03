#!/bin/bash
#
# This script is run by systemd using the ExecStartPre option
# before starting bitcoind.service (Bitcoin Core).
#
set -eu

# check if SSD is already available to avoid failure
BITCOIN_DIR="/mnt/ssd/bitcoin/.bitcoin"
if [ ! -d "${BITCOIN_DIR}" ] || [ ! -x "${BITCOIN_DIR}" ]; then
    echo "ERR: cannot start 'bitcoind', directory ${BITCOIN_DIR} not accessible"
    exit 1
else
    echo "INFO: starting 'bitcoind', directory ${BITCOIN_DIR} accessible"
fi
