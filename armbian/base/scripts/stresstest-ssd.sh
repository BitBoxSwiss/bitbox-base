#!/bin/bash
set -eu

# BitBoxBase: stresstest the ssd with stressdisk utility
#
# prerequisites: installed tmux and unzip

function usage() {
  echo "BitBoxBase: stresstest the ssd
usage: stresstest-ssd.sh
"
}

if [[ ${#} -gt 0 ]]; then
  usage
  exit 1
fi

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

if which stressdisk; then
    cd /tmp
    wget https://github.com/ncw/stressdisk/releases/download/v1.0.12/stressdisk_1.0.12_linux_arm64.zip
    unzip stressdisk_1.0.12_linux_arm64.zip
    chmod +x stressdisk
    mv stressdisk /usr/sbin
fi

mkdir -p /mnt/ssd/stressdisk

tmux new-session -d 'watch smartctl -a /dev/nvme0n1'
tmux split-window -h 'stressdisk cycle /mnt/ssd/stressdisk'
tmux split-window -v 'htop'
tmux split-window -v 'bbbfancontrol -v'
tmux -2 attach-session -d
