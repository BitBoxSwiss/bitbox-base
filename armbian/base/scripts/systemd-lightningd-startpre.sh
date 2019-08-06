#!/bin/bash
#
# This script is run by systemd using the ExecStartPre option 
# before starting lightningd.service (c-lightning).
#

if ! systemctl is-active bitcoind.service; then
    echo "${0}: startup checks failed. bitcoind.service is not active. Not starting lightningd.service."
    exit 1
fi

if [ ! -f /data/triggers/bitcoind_fully_synced ]; then
    echo "${0}: startup checks failed. File /data/triggers/bitcoind_fully_synced not present, thus bitcoind not fully synced. Not starting lightningd.service."
    exit 1
fi

if [ -f /mnt/ssd/bitcoin/.bitcoin/.cookie ]; then
    echo -n 'RPCPASSWORD=' > /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    tail -c +12 /mnt/ssd/bitcoin/.bitcoin/.cookie >> /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    chown bitcoin:bitcoin /mnt/ssd/bitcoin/.bitcoin/.cookie.env
    echo "${0}: file /mnt/ssd/bitcoin/.bitcoin/.cookie.env updated."
else
    echo "${0}: startup checks failed. Authentication file /mnt/ssd/bitcoin/.bitcoin/.cookie not present, not starting lightningd.service."
    exit 1
fi