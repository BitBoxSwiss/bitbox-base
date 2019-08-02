#!/bin/bash
#
# This script is run before starting electrs.service (Electrum Server in Rust).
#

# load configuration file
source /etc/electrs/electrs.conf
source /mnt/ssd/bitcoin/.bitcoin/.cookie.env

if ! systemctl is-active bitcoind.service; then
    echo "systemd-electrs-startchecks.sh failed: bitcoind.service is not active. Not starting electrs.service."
    exit 1
fi

if [ ! -f /data/triggers/bitcoind_fully_synced ]; then
    echo "systemd-electrs-startchecks.sh failed: file /data/triggers/bitcoind_fully_synced not present, thus bitcoind not fully synced. Not starting electrs.service."
    exit 1
fi

if [ -f /mnt/ssd/bitcoin/.bitcoin/.cookie ]; then
    echo -n 'RPCPASSWORD=' > /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    tail -c +12 /mnt/ssd/bitcoin/.bitcoin/.cookie >> /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    chown bitcoin:bitcoin /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    echo "systemd-electrs-startchecks.sh: file /mnt/ssd/bitcoin/.bitcoin/.cookie.env updated."
else
    echo "systemd-electrs-startchecks.sh failed: authentication file /mnt/ssd/bitcoin/.bitcoin/.cookie not present, not starting electrs.service."
    exit 1
fi