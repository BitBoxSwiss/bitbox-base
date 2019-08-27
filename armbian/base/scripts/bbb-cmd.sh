#!/bin/bash
set -eux

# BitBox Base: system commands repository
#

# print usage information for script
usage() {
  echo "BitBox Base: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  setup         <datadir|hostname_link>
  base          <restart|shutdown>
  bitcoind      <reindex|resync>
  flashdrive    <check|mount|umount>
  backup        <sysconfig|hsm_secret>
  restore       <sysconfig|hsm_secret>
                    
"
}

# function to execute command, either within overlayroot-chroot or directly
exec_overlayroot() {
    if [[ "${1}" != "base-only" ]] && [[ "${1}" != "all-layers" ]]; then
        echo "exec_overlayroot(): first argument '${1}', but must be either"
        echo "                    'base-only':  execute base layer (in r/o partition when overlayroot active, or directy when no overlayroot active"
        echo "                    'all-layers': execute both in overlayroot and directly"
        exit 1
    fi

    if [ "${OVERLAYROOT_ENABLED}" -eq 1 ]; then
        echo "executing in overlayroot-chroot: ${2}"
        overlayroot-chroot /bin/bash -c "${2}"
    fi

    if [ "${OVERLAYROOT_ENABLED}" -ne 1 ] || [[ "${1}" == "all-layers" ]]; then
        echo "executing directly: ${2}"
        /bin/bash -c "${2}"
    fi
}

redis_set() {
    # usage: redis_set "key" "value"
    ok=$(redis-cli -h localhost -p 6379 -n 0 SET "${1}" "${2}")
    if [[ "${ok}"  != "OK" ]]; then
        echo "ERR: could not SET key ${1}"
        # exit 1
    fi
}

redis_get() {
    # usage: str=$(redis_get "key")
    ok=$(redis-cli -h localhost -p 6379 -n 0 GET "${1}")
    echo "${ok}"
}

generateConfig() {
  # generates a config file using custom bbbconfig
  #
  # argument is template filename, without path
  #
  local TEMPLATES_DIR="/opt/shift/config/templates"

  if [ ${#} -eq 0 ] || [ ${#} -gt 1 ]; then
    echo "ERR: generateConfig() expects exactly one argument"
    exit 1
  fi

  local FILE="${TEMPLATES_DIR}/${1}"
  if [ -f "${FILE}" ]; then
    echo "generateConfig() from ${FILE}"
    /usr/local/sbin/bbbconfgen --template "${FILE}"
  else
    echo "ERR: generateConfig() template file ${FILE} not found"
    exit 1
  fi
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

# check if overlayroot is enabled
OVERLAYROOT_ENABLED=0
if grep -q "tmpfs" /etc/overlayroot.local.conf; then
    OVERLAYROOT_ENABLED=1
fi

MODULE="${1:-''}"
COMMAND="${2:-''}"
ARG="${3:-''}"

MODULE="${MODULE^^}"
COMMAND="${COMMAND^^}"

case "${MODULE}" in
    SETUP)
        case "${COMMAND}" in
            DATADIR)

                echo "check"
                if [ ! -f /data/.datadir_set_up ]; then
                    # if /data is separate partition, data is copied (assume /data is read/write)
                    echo "mountpoint"
                    if mountpoint /data -q; then
                        cp -r /data_source/* /data
                    
                    # otherwise create symlink
                    else
                        echo "ln"
                        ln -sf /data_source /data
                    fi
                    echo "OK: (DATADIR) symlink /data/ --> /data_source/ created"

                else
                    echo "WARN: (DATADIR) data directory already set up (found file /data/.datadir_set_up)"
                fi
                ;;

            DATADIR_OVERLAY)
                # this command needs to be executed within the overlayroot base layer
                # e.g. by running 'overlayroot-chroot /bin/bash -c "/opt/shift/scripts/bbb-cmd.sh setup datadir_overlay"
                if [ ! -f /data/.datadir_set_up ]; then
                    ln -sf /data_source /data
                    echo "OK: (DATADIR_OVERLAY) symlink /data/ --> /data_source/ created"
                else
                    echo "WARN: (DATADIR_OVERLAY) data directory already set up (found file /data/.datadir_set_up)"
                fi
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;

    BITCOIND)
        case "${COMMAND}" in
            RESYNC|REINDEX)
                # stop systemd services
                systemctl stop electrs
                systemctl stop lightningd

                # deleting bitcoind chainstate in /mnt/ssd/bitcoin/.bitcoin/chainstate
                rm -rf /mnt/ssd/bitcoin/.bitcoin/chainstate
                rm -rf /mnt/ssd/electrs/db

                # for RESYNC incl. download, delete `blocks` directory too
                if [[ "${COMMAND}" == "RESYNC" ]]; then
                    rm -rf /mnt/ssd/bitcoin/.bitcoin/blocks
                    echo "X"
                fi

                redis_set "bitcoind:ibd" "1"
                redis_set "bitcoind:reindex-chainstate" 1
                generateConfig "bitcoin.conf.template"
                sleep 5

                # restart bitcoind and remove option
                systemctl start bitcoind.service
                sleep 10
                
                redis_set "bitcoind:reindex-chainstate" 0
                generateConfig "bitcoin.conf.template"

                echo "Command ${MODULE} ${COMMAND} successfully executed."
                ;;
            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
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

    FLASHDRIVE)
        # sanitize device name
        ARG="${ARG//[^a-zA-Z0-9_\/]/}"

        case "${COMMAND}" in
            CHECK)
                flashdrive_count=0
                while read -r scsidev; do
                    name=$( echo "${scsidev}" | cut -s -f 1 -d " " )
                    size=$( echo "${scsidev}" | cut -s -f 2 -d " " )
                    fstype=$( echo "${scsidev}" | cut -s -f 3 -d " " )

                    # search drives that have a filesystem (not iso9660) and are <= 64GB
                    if [ -n "${fstype}" ] && [[ "${fstype}" != "iso9660" ]] && [[ size -lt 64000000000 ]]; then
                        flashdrive_count=$((flashdrive_count + 1))
                        flashdrive_name="${name}"
                    fi
                done <<< "$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8)"

                # only 1 drive must be present, otherwise abort
                if [[ flashdrive_count -eq 1 ]]; then
                    echo "${flashdrive_name}"

                elif [[ flashdrive_count -eq 0 ]]; then
                    echo "FLASHDRIVE CHECK: no target found"
                    exit 1

                else
                    echo "FLASHDRIVE CHECK: too many targets found (${flashdrive_count} in total)"
                    exit 1
                fi               
                ;;

            MOUNT)
                # ensure mountpoint is available
                mkdir -p /mnt/backup

                # check if ARG is valid flashdrive
                if ! lsblk "${ARG}" > /dev/null 2>&1; then
                    echo "FLASHDRIVE MOUNT: device ${ARG} not found."
                    exit 1

                elif [ "$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8 "${ARG}" | wc -l)" -ne 1 ]; then
                    echo "FLASHDRIVE MOUNT: device ${ARG} is not unique and/or has partitions."
                    exit 1

                else
                    scsidev=$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8 "${ARG}")
                    name=$( echo "${scsidev}" | cut -s -f 1 -d " " )
                    size=$( echo "${scsidev}" | cut -s -f 2 -d " " )
                    fstype=$( echo "${scsidev}" | cut -s -f 3 -d " " )

                    if [ -n "${fstype}" ] && [[ "${fstype}" != "iso9660" ]] && [[ size -lt 64000000000 ]]; then

                        # mount usb flashdrive with the following options:
                        #   rw:         read/write
                        #   nosuid:     cannot contain set userid files, prevents root escalation
                        #   nodev:      cannot contain special devices
                        #   noexec:     cannot contain executable binaries
                        #   noatime:    no update of file access time
                        #   nodiratime: no update of directory access time
                        #   sync:       synchronous write, flushed to disk immediatly
                        mount "${ARG}" -o rw,nosuid,nodev,noexec,noatime,nodiratime,sync /mnt/backup
                        echo "FLASHDRIVE MOUNT: mounted ${name} to /mnt/backup"

                    else
                        echo "FLASHDRIVE MOUNT: device ${name} is either bigger than 64GB (${size}) or does the filesystem (${fstype}) is not supported."
                        exit 1
                    fi

                fi
                ;;

            UNMOUNT)
                if ! mountpoint /mnt/backup -q; then
                    echo "FLASHDRIVE UNMOUNT: no drive mounted at /mnt/backup"
                    exit 1
                else
                    umount /mnt/backup
                    echo "FLASHDRIVE UNMOUNT: /mnt/backup unmounted."
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
