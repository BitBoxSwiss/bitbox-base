#!/bin/bash
set -eu

# BitBox Base: system commands repository
#

# print usage information for script
function usage() {
  echo "BitBox Base: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  bitcoind reindex  wipes UTXO set and validates existing blocks
  bitcoind resync   re-download and validate all blocks

  base restart      restarts the node
  base shutdown     shuts down the node
  
  usb_flashdrive    <check|mount|umount>
  backup            <sysconfig|hsm_secret>
  restore           <sysconfig|hsm_secret>
                    
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
                exit 1
        esac
        ;;

    USB_FLASHDRIVE)
        # sanitize device name
        ARG="${ARG//[^a-zA-Z0-9_\/]/}"

        case "${COMMAND}" in
            CHECK)
                usb_flashdrive_count=0
                while read -r scsidev; do
                    name=$( echo "${scsidev}" | cut -s -f 1 -d " " )
                    size=$( echo "${scsidev}" | cut -s -f 2 -d " " )
                    fstype=$( echo "${scsidev}" | cut -s -f 3 -d " " )

                    # search drives that have a filesystem (not iso9660) and are <= 64GB
                    if [ -n "${fstype}" ] && [[ "${fstype}" != "iso9660" ]] && [[ size -lt 64000000000 ]]; then
                        usb_flashdrive_count=$((usb_flashdrive_count + 1))
                        usb_flashdrive_name="${name}"
                    fi
                done <<< "$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8)"

                # only 1 drive must be present, otherwise abort
                if [[ usb_flashdrive_count -eq 1 ]]; then
                    echo "${usb_flashdrive_name}"

                elif [[ usb_flashdrive_count -eq 0 ]]; then
                    echo "USB_FLASHDRIVE CHECK: no target found"
                    exit 1

                else
                    echo "USB_FLASHDRIVE CHECK: too many targets found (${usb_flashdrive_count} in total)"
                    exit 1
                fi               
                ;;

            MOUNT)
                # ensure mountpoint is available
                mkdir -p /mnt/backup

                # check if ARG is valid USB flashdrive
                if ! lsblk "${ARG}" > /dev/null 2>&1; then
                    echo "USB_FLASHDRIVE MOUNT: device ${ARG} not found."
                    exit 1

                elif [ "$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8 "${ARG}" | wc -l)" -ne 1 ]; then
                    echo "USB_FLASHDRIVE MOUNT: device ${ARG} is not unique and/or has partitions."
                    exit 1

                else
                    scsidev=$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8 "${ARG}")
                    name=$( echo "${scsidev}" | cut -s -f 1 -d " " )
                    size=$( echo "${scsidev}" | cut -s -f 2 -d " " )
                    fstype=$( echo "${scsidev}" | cut -s -f 3 -d " " )

                    if [ -n "${fstype}" ] && [[ "${fstype}" != "iso9660" ]] && [[ size -lt 64000000000 ]]; then

                        # mount USB device with the following options:
                        #   rw:         read/write
                        #   nosuid:     cannot contain set userid files, prevents root escalation
                        #   nodev:      cannot contain special devices
                        #   noexec:     cannot contain executable binaries
                        #   noatime:    no update of file access time
                        #   nodiratime: no update of directory access time
                        #   sync:       synchronous write, flushed to disk immediatly
                        mount "${ARG}" -o rw,nosuid,nodev,noexec,noatime,nodiratime,sync /mnt/backup
                        echo "USB_FLASHDRIVE MOUNT: mounted ${name} to /mnt/backup"

                    else
                        echo "USB_FLASHDRIVE MOUNT: device ${name} is either bigger than 64GB (${size}) or does the filesystem (${fstype}) is not supported."
                        exit 1
                    fi

                fi
                ;;

            UNMOUNT)
                if ! mountpoint /mnt/backup -q; then
                    echo "USB_FLASHDRIVE UNMOUNT: no drive mounted at /mnt/backup"
                    exit 1
                else
                    umount /mnt/backup
                    echo "USB_FLASHDRIVE UNMOUNT: /mnt/backup unmounted."
                fi
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;        
    
    BACKUP)
        REDIS_FILEPATH="/data/redis/bitboxbase.rdb"
        HSM_FILEPATH="/mnt/ssd/bitcoin/.lightning/hsm_secret"

        case "${COMMAND}" in
            # backup system configuration to mounted usb flashdrive
            SYSCONFIG)
                if mountpoint /mnt/backup -q; then
                    cp "${REDIS_FILEPATH}" "/mnt/backup/bbb-backup.rdb"

                    # create backup history (restore not yet implemented)
                    cp "${REDIS_FILEPATH}" "/mnt/backup/bbb-backup_$(date '+%Y%m%d-%H%M').rdb"
                else
                    echo "ERR: /mnt/backup is not a mountpoint"
                    exit 1
                fi
                echo "OK: backup created as /mnt/backup/bbb-backup.rdb"
                ;;

            # backup c-lightning on-chain keys in 'hsm_secret' into Redis database
            HSM_SECRET)
                # encode binary file 'hsm_secret' as base64 and store it in Redis
                redis-cli SET lightningd:hsm_secret "$(base64 < ${HSM_FILEPATH})"
                echo "OK: backup of file 'hsm_secret' created"
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;    

    RESTORE)
        REDIS_FILEPATH="/data/redis/bitboxbase.rdb"
        HSM_FILEPATH="/mnt/ssd/bitcoin/.lightning/hsm_secret"

        case "${COMMAND}" in
            # restore system configuration from mounted usb flashdrive
            SYSCONFIG)
                if [ -f /mnt/backup/bbb-backup.rdb ]; then
                    cp "/mnt/backup/bbb-backup.rdb" "${REDIS_FILEPATH}"
                    systemctl restart redis-server.service
                else
                    echo "ERR: backup file /mnt/backup/bbb-backup.rdb not found"
                    exit 1
                fi
                echo "OK: backup file /mnt/backup/bbb-backup.rdb restored to ${REDIS_FILEPATH}."
                ;;

            # restore c-lightning on-chain keys from Redis database
            HSM_SECRET)
                # create snapshot of 'hsm_secret'
                if [ -f "${HSM_FILEPATH}" ]; then
                    cp "${HSM_FILEPATH}" "${HSM_FILEPATH}_$(date '+%Y%m%d-%H%M').backup"
                else
                    echo "WARN: no previous 'hsm_secret' found, no local backup created"
                fi

                # save base64 encoded 'hsm_secret' as binary file to file system
                # redis-cli causes script to terminate when Redis not available
                redis-cli GET lightningd:hsm_secret | base64 -d > /mnt/ssd/bitcoin/.lightning/hsm_secret
                echo "OK: backup of file 'hsm_secret' restored"
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;    

    *)
        echo "Invalid argument: module ${MODULE} unknown."
        exit 1
esac
