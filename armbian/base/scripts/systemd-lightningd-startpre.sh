#!/bin/bash
# shellcheck disable=SC1091
#
# This script is run by systemd using the ExecStartPre option
# before starting lightningd.service (c-lightning).
#

set -eu

# --- generic functions --------------------------------------------------------

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# include errorExit() function
source /opt/shift/scripts/include/errorExit.sh.inc

# ------------------------------------------------------------------------------

if ! systemctl is-active bitcoind.service; then
    echo "ERR: bitcoind.service is not active. Not starting lightningd.service."
    errorExit BITCOIND_DEPENDENCY_NOT_ACTIVE
fi

# check if bitcoind is in Initial Block Download (IBD) mode
BITCOIN_IBD=$(redis_get 'bitcoind:ibd')
BITCOIN_IBD=${BITCOIN_IBD:-0}

if [ "${BITCOIN_IBD}" -eq 1 ]; then
    echo "ERR: bitcoind.service is in IBD mode. Not starting lightningd.service."
    errorExit BITCOIND_IN_IBD
fi
