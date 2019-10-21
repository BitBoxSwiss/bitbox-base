#!/bin/bash
# shellcheck disable=SC1091
#
# This script is run by systemd using the ExecStartPre option
# before starting bbbmiddleware.service.
#
set -eu

# --- generic functions --------------------------------------------------------

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# ------------------------------------------------------------------------------

REFRESH_RPCAUTH="$(redis_get 'bitcoind:refresh-rpcauth')"
if [ "${REFRESH_RPCAUTH}" -ne 0 ]; then
    # either Redis not ready yet or new credentials requested
    echo "INFO: bitcoind:refresh-rpcauth not 0, holding off bbbmiddleware start for bitcoind to warm up"
    sleep 15

else
    echo "INFO: bitcoind:refresh-rpcauth equals 0, starting bbbmiddleware immediately"
fi
