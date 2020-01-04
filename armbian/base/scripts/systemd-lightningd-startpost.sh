#!/bin/bash
# shellcheck disable=SC1091
#
# This script is run by systemd using the ExecStartPost option
# after starting lightningd.service (c-lightning).
#

set -eu

# --- generic functions --------------------------------------------------------

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# ------------------------------------------------------------------------------

# wait for c-lightning to warm up
sleep 5

# make available lightningd socket to group "bitcoin"
chown -R bitcoin:bitcoin /mnt/ssd/bitcoin/.lightning
chmod -R u+rw,g+rx,g-w,o-rwx /mnt/ssd/bitcoin/.lightning

chmod 700 /mnt/ssd/bitcoin/.lightning/bitcoin/* || true
chmod 770 /mnt/ssd/bitcoin/.lightning/bitcoin/lightning-rpc || true

chmod 700 /mnt/ssd/bitcoin/.lightning/testnet/* || true
chmod 770 /mnt/ssd/bitcoin/.lightning/testnet/lightning-rpc || true

# update tor address in Redis
redis_set "tor:lightningd:onion" "$(lightning-cli --conf=/etc/lightningd/lightningd.conf getinfo | jq -r '.address[0] .address')"
