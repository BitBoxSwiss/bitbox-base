#!/bin/bash
#
# This script is called by the lightningd.service AFTER starting c-lightning.
#

# make available lightningd socket to group "bitcoin"
source /opt/shift/sysconfig/BITCOIN_NETWORK

if [[ "${BITCOIN_NETWORK}" == "mainnet" ]]; then
    chmod g+rwx /mnt/ssd/bitcoin/.lightning/lightning-rpc
else
    chmod g+rwx /mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc
fi
