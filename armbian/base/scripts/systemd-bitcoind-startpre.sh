#!/bin/bash
#
# This script is run by systemd using the ExecStartPre option
# before starting bitcoind.service (Bitcoin Core).
#
# Must be run as 'root' with ExecStartPre=+
#
set -eu

# include functions redis_set() and redis_get()
# shellcheck disable=SC1091
source /opt/shift/scripts/include/redis.sh.inc

# ------------------------------------------------------------------------------

# Redis must be available
redis_require

# check if rpcauth credentials exist, or create new ones
RPCAUTH="$(redis_get 'bitcoind:rpcauth')"
REFRESH_RPCAUTH="$(redis_get 'bitcoind:refresh-rpcauth')"

if [ ${#RPCAUTH} -lt 90 ] || [ "${REFRESH_RPCAUTH}" -eq 1 ]; then
    echo "INFO: creating new bitcoind rpc credentials"
    echo "INFO: old bitcoind:rpcauth was ${RPCAUTH}"
    echo "INFO: bitcoind:refresh-rpcauth is ${REFRESH_RPCAUTH}"
    /opt/shift/scripts/bbb-cmd.sh bitcoind refresh_rpcauth
else
    echo "INFO: found bitcoind rpc credentials, no action taken"
fi

# check if SSD is already available to avoid failure
BITCOIN_DIR="/mnt/ssd/bitcoin/.bitcoin"
if [ ! -d "${BITCOIN_DIR}" ] || [ ! -x "${BITCOIN_DIR}" ]; then
    echo "ERR: cannot start 'bitcoind', directory ${BITCOIN_DIR} not accessible"
    exit 1
else
    echo "INFO: starting 'bitcoind', directory ${BITCOIN_DIR} accessible"
fi
