#!/bin/bash
#
# create BitBoxBase presync archive
#
set -eu

# print usage information for script
usage() {
  echo "BitBoxBase: create presync archive
usage: create-presync-archive.sh [--help]

This script needs to be run on the BitBoxBase and creates an uncompressed archive
with Bitcoin Core blocks/chainstate and Electrs database for presynced devices.

Requirements:
- Bitcoin Core datadir at /mnt/ssd/bitcoin/.bitcoin
- Electrs database at     /mnt/ssd/electrs/db/mainnet

The presync archive will be created as /mnt/ssd/bbb-presync-ssd-YYYYMMDD-hhmm.tar

"
}

if [[ ${#} -ne 0 ]]; then
    usage
    exit 0
fi

if [ ! -d /mnt/ssd/bitcoin/.bitcoin ] || [ ! -d /mnt/ssd/electrs/db/mainnet ]; then
    echo "ERR: required directories not found (run with --help for additional details)" >&2
    exit 1
fi

if [[ ${UID} -ne 0 ]]; then
    echo "ERR: needs to be run as superuser." >&2
    exit 1
fi

freespace=$(df -k /mnt/ssd  | awk '/[0-9]%/{print $(NF-2)}')
if [[ ${freespace} -lt 400000000 ]]; then
    echo "ERR: not enough disk space, should at least have 400 GB" >&2
    exit 1
fi

cd /mnt/ssd || exit

tar cvfW bbb-presync-ssd-"$(date '+%Y%m%d-%H%M')".tar \
    bitcoin/.bitcoin/blocks \
    bitcoin/.bitcoin/chainstate \
    --exclude='IDENTITY' \
    --exclude='LOG*' \
    --exclude='*.log' \
    electrs/db/mainnet

echo
echo "Archive created:"
echo
ls -lh /mnt/ssd/bbb-presync*
echo
