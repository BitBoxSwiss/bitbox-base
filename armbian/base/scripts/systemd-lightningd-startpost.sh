#!/bin/bash
#
# This script is run by systemd using the ExecStartPost option 
# after starting lightningd.service (c-lightning).
#

set -eu

redis_get() {
    # usage: str=$(redis_get "key")
    ok=$(redis-cli -h localhost -p 6379 -n 0 GET "${1}")
    echo "${ok}"
}

# ------------------------------------------------------------------------------

BITCOIN_NETWORK=$(redis_get "bitcoind:network")

# wait for c-lightning to warm up
sleep 10

# make available lightningd socket to group "bitcoin"
if [[ "${BITCOIN_NETWORK}" == "mainnet" ]]; then
    chmod g+rwx /mnt/ssd/bitcoin/.lightning/lightning-rpc
elif [ -d /mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc ]; then
    chmod g+rwx /mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc
else
    echo "Failed to set permissions to lightning-rpc socket."
fi
