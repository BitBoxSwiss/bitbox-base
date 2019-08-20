#!/bin/bash
set -eu

# BitBox Base: system commands repository
#

# print usage information for script
function usage() {
  echo "BitBox Base: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  bitcoind reindex   wipes UTXO set and validates existing blocks
  bitcoind resync    re-download and validate all blocks

  base restart      restarts the node
  base shutdown     shuts down the node
  
  usb_thumbdrive    <check|mount|umount>
  usb_backup

"
}

# ------------------------------------------------------------------------------

# check script arguments
if [[ ${#} == 0 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
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

MODULE="${1:-''}"
COMMAND="${2:-''}"
ARG="${3:-''}"

MODULE="${MODULE^^}"
COMMAND="${COMMAND^^}"

case "${MODULE}" in
    BITCOIND)
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
            if [[ "${COMMAND}" == "RESYNC" ]]; then
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

        echo "Command ${MODULE} ${COMMAND} successfully executed."
        ;;

    BASE)
        case "${COMMAND}" in
            RESTART)
                reboot
                ;;
            SHUTDOWN)
                shutdown now
                ;;
            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
        esac
        ;;

    USB_THUMBDRIVE)
        # sanitize device name
        ARG="${ARG//[^a-zA-Z0-9_\/]/}"

        case "${COMMAND}" in
            CHECK)
                usb_thumbdrive_count=0
                while read scsidev; do
                    name=$( echo "${scsidev}" | cut -f 1 -d " " )
                    size=$( echo "${scsidev}" | cut -f 2 -d " " )
                    type=$( echo "${scsidev}" | cut -f 3 -d " " )

                    # search drives that have partition and are <= 64GB
                    if [[ "${type}" == "part" ]] && [[ size -lt 64000000000 ]]; then
                        usb_thumbdrive_count=$((usb_thumbdrive_count + 1))
                        usb_thumbdrive_name="${name}"
                    fi
                done <<< "$(lsblk -o NAME,SIZE,TYPE -abrnp -I 8)"

                # only 1 drive must be present, otherwise abort
                if [[ usb_thumbdrive_count -eq 1 ]]; then
                    echo "${usb_thumbdrive_name}"

                elif [[ usb_thumbdrive_count -eq 0 ]]; then
                    echo "USB_THUMBDRIVE CHECK: no target found"
                    exit 1

                else
                    echo "USB_THUMBDRIVE CHECK: too many targets found (${usb_thumbdrive_count} in total)"
                    exit 1
                fi               
                ;;

            MOUNT)
                # check if ARG is valid USB thumbdrive
                if ! lsblk "${ARG}" > /dev/null 2>&1; then
                    echo "USB_THUMBDRIVE MOUNT: device ${ARG} not found."
                    exit 1

                elif [ `lsblk -o NAME,SIZE,TYPE -abrnp -I 8 ${ARG} | wc -l` -ne 1 ]; then
                    echo "USB_THUMBDRIVE MOUNT: device ${ARG} is not unique and/or has partitions."
                    exit 1

                else
                    scsidev=`lsblk -o NAME,SIZE,TYPE -abrnp -I 8 ${ARG}`
                    size=$( echo "${scsidev}" | cut -f 2 -d " " )
                    type=$( echo "${scsidev}" | cut -f 3 -d " " )

                    if [[ "${type}" != "part" ]] || [[ size -gt 64000000000 ]]; then
                        echo "USB_THUMBDRIVE MOUNT: device ${scsidev} is either bigger than 64GB (${size}) or not a partition (type ${type})."
                        exit 1

                    # all checks passed
                    else
                        # mount USB device with the following options:
                        #   rw:         read/write
                        #   nosuid:     cannot contain set userid files, prevents root escalation
                        #   nodev:      cannot contain special devices
                        #   noexec:     cannot contain executable binaries
                        #   noatime:    no update of file access time
                        #   nodiratime: no update of directory access time
                        #   sync:       synchronous write, flushed to disk immediatly
                        mount "${ARG}" -o rw,nosuid,nodev,noexec,noatime,nodiratime,sync /mnt/backup
                        echo "USB_THUMBDRIVE MOUNT: mounted ${scsidev}} to /mnt/backup"
                    fi

                fi
                ;;

            UNMOUNT)
                if ! mountpoint /mnt/backup -q; then
                    echo "USB_THUMBDRIVE UNMOUNT: no drive mounted at /mnt/backup"
                else
                    echo "USB_THUMBDRIVE UNMOUNT: /mnt/backup unmounted."
                    umount /mnt/backup
                fi
                ;;
            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
        esac
        ;;        
    *)
        echo "Invalid argument: module ${MODULE} unknown."

esac
