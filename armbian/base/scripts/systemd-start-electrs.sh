#!/bin/bash
#
# This script is called by the electrs.service to start electrs (Electrum Server in Rust).
#

# load configuration file
source /etc/electrs/electrs.conf
source /mnt/ssd/bitcoin/.bitcoin/.cookie.env

# start main application
/usr/bin/electrs \
    --network ${NETWORK} \
    --db-dir ${DB_DIR} \
    --daemon-dir ${DAEMON_DIR} \
    --cookie "__cookie__:${RPCPASSWORD}" \
    --monitoring-addr ${MONITORING_ADDR} \
    -${VERBOSITY}
