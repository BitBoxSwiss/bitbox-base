#!/bin/bash
#
# This script is run by systemd using the ExecStartPre option 
# before starting electrs.service (Electrum Server in Rust).
#

if ! systemctl is-active bitcoind.service; then
    echo "${0} startup checks failed. bitcoind.service is not active. Not starting electrs.service."
    exit 1
fi

if [ ! -f /data/triggers/bitcoind_fully_synced ]; then
    echo "${0} startup checks failed. File /data/triggers/bitcoind_fully_synced not present, thus bitcoind not fully synced. Not starting electrs.service."
    exit 1
fi

if [ -f /mnt/ssd/bitcoin/.bitcoin/.cookie ]; then
    sleep 3
    echo -n 'RPCPASSWORD=' > /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    tail -c +12 /mnt/ssd/bitcoin/.bitcoin/.cookie >> /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    chown bitcoin:bitcoin /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    echo "${0}: file /mnt/ssd/bitcoin/.bitcoin/.cookie.env updated."
else
    echo "${0} startup checks failed. Authentication file /mnt/ssd/bitcoin/.bitcoin/.cookie not present, not starting electrs.service."
    exit 1
fi
