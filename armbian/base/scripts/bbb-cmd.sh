#!/bin/bash
set -eu

# BitBox Base: system commands repository
#

# print usage information for script
function usage() {
  echo "BitBox Base: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  bitcoin_reindex   wipes UTXO set and validates existing blocks
  bitcoin_resync    re-download and validate all blocks

  base_restart      restarts the node
  base_shutdown     shuts down the node

"
}

# ------------------------------------------------------------------------------

# check script arguments
if [[ ${#} -ne 1 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
  usage
  exit 0
elif [[ "${1}" == "-v" ]] || [[ "${1}" == "--version" ]]; then
  echo "bbb-cmd version 0.1"
  exit 0
fi

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

COMMAND="${1^^}"

case "${COMMAND}" in
    BITCOIN_REINDEX|BITCOIN_RESYNC)
        # stop systemd services
        systemctl stop electrs
        systemctl stop lightningd
        systemctl stop bitcoind

        if ! /bin/systemctl -q is-active bitcoind.service; then 
            # deleting bitcoind chainstate in /mnt/ssd/bitcoin/.bitcoin/chainstate
            rm -rf /mnt/ssd/bitcoin/.bitcoin/chainstate
            rm -rf /mnt/ssd/electrs/db
            rm -rf /data/triggers/bitcoind_fully_synced

            # for RESYNC incl. download, delete `blocks` directory too
            if [[ "${COMMAND}" == "BITCOIN_RESYNC" ]]; then
                rm -rf /mnt/ssd/bitcoin/.bitcoin/blocks

            # otherwise assume REINDEX (only validation, no download), set option reindex-chainstate
            else
                echo "reindex-chainstate=1" >> /etc/bitcoin/bitcoin.conf

            fi

            # restart bitcoind and remove option
            systemctl start bitcoind
            sleep 10
            sed -i '/reindex/Id' /etc/bitcoin/bitcoin.conf

        else
            echo "bitcoind is still running, cannot delete chainstate"
            exit 1
        fi

        echo "Command ${COMMAND} successfully executed."
        ;;

    BASE_RESTART)
        reboot
        ;;

    BASE_SHUTDOWN)
        shutdown now
        ;;
        
    *)
        echo "Invalid argument: command ${COMMAND} unknown."

esac
