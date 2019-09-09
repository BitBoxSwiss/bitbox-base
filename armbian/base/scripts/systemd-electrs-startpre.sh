#!/bin/bash
#
# This script is run by systemd using the ExecStartPre option 
# before starting electrs.service (Electrum Server in Rust).
#

set -eu

redis_get() {
    # usage: str=$(redis_get "key")
    ok=$(redis-cli -h localhost -p 6379 -n 0 GET "${1}")
    echo "${ok}"
}

# ------------------------------------------------------------------------------

if ! systemctl is-active bitcoind.service; then
    echo "ERR: bitcoind.service is not active. Not starting electrs.service."
    exit 1
fi

# check if bitcoind is in Initial Block Download (IBD) mode
BITCOIN_IBD=$(redis_get 'bitcoind:ibd')
BITCOIN_IBD=${BITCOIN_IBD:-0}

if [ $BITCOIN_IBD -eq 1 ]; then
    echo "ERR: bitcoind.service is in IBD mode. Not starting electrs.service."
    exit 1
fi
