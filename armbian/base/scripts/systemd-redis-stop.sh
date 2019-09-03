#!/bin/bash
#
# This script is run by systemd using the ExecStop option
# when stopping redis.service.
#

set -eu
if [ -d /mnt/ssd/data/redis ]; then
    cp --update "/data/redis/bitboxbase.rdb" "/mnt/ssd/data/redis/bitboxbase.backup.rdb"
    cp --update "/data/redis/bitboxbase.rdb" "/mnt/ssd/data/redis/bitboxbase.$(date '+%Y%m%d-%H%M').rdb"
    echo "INFO: created backup of Redis configuration store in /mnt/ssd/data/redis/"
else
    echo "ERR: backup directory /mnt/ssd/data/redis/ not found, cannot create backup of Redis configuratoin store"
    exit 1
fi
