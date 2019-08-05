#!/bin/bash
#
# This script is run by systemd using the ExecStartPost option 
# after starting lightningd.service (c-lightning).
#

# make available lightningd socket to group "bitcoin"
source /data/sysconfig/BITCOIN_NETWORK

# wait for c-lightning to warm up
sleep 10

if [[ "${BITCOIN_NETWORK}" == "mainnet" ]]; then
    chmod g+rwx /mnt/ssd/bitcoin/.lightning/lightning-rpc
elif [ -d /mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc ]; then
    chmod g+rwx /mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc
else
    echo "Failed to set permissions to lightning-rpc socket."
fi
