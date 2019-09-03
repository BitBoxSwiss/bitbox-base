#!/bin/bash
#
# This script is run by systemd using the ExecStartPost option
# before starting redis.service.
set -eux

# if overlayroot is enabled, copy database from ssd
OVERLAYROOT_ENABLED=0
if grep -q "tmpfs" /etc/overlayroot.local.conf; then
    if [ -f /mnt/ssd/data/redis/bitboxbase.backup.rdb ]; then
        cp --update /mnt/ssd/data/redis/bitboxbase.backup.rdb /data/redis/bitboxbase.rdb
    else
        echo "WARN: overlayrootfs detected, but no persistent copy of Redis database found at /mnt/ssd/data/redis/bitboxbase.backup.rdb"
    fi
else
    echo "INFO: no overlayrootfs detected, Redis backup is not restored from /mnt/ssd/data/redis/bitboxbase.backup.rdb"
fi
