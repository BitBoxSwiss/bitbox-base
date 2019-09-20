#!/bin/bash
# shellcheck disable=SC1091
#
# This script is run by systemd using the ExecStartPost option 
# after starting lightningd.service (c-lightning).
#

set -eu

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

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
