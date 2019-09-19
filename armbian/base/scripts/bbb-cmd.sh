#!/bin/bash
# shellcheck disable=SC1091
set -eu

# BitBox Base: system commands repository
#

# print usage information for script
usage() {
  echo "BitBox Base: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  setup         <datadir>
  base          <restart|shutdown>
  bitcoind      <reindex|resync|refresh_rpcauth>
  flashdrive    <check|mount|umount>
  backup        <sysconfig|hsm_secret>
  restore       <sysconfig|hsm_secret>
  mender-update <install|commit>

"
}

# include function exec_overlayroot(), to execute a command, either within overlayroot-chroot or directly
source /opt/shift/scripts/include/exec_overlayroot.sh.inc

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# include function generateConfig() to generate config files from templates
source /opt/shift/scripts/include/generateConfig.sh.inc

# include errorExit() function
source /opt/shift/scripts/include/errorExit.sh.inc

# ------------------------------------------------------------------------------

# check script arguments
if [[ ${#} -lt 2 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
    usage
    exit 0
elif [[ "${1}" == "-v" ]] || [[ "${1}" == "--version" ]]; then
    echo "bbb-cmd version 0.1"
    exit 0
fi

if [[ ${UID} -ne 0 ]]; then
    echo "${0}: needs to be run as superuser."
    errorExit SCRIPT_NOT_RUN_AS_SUPERUSER
fi

MODULE="${1:-}"
COMMAND="${2:-}"
ARG="${3:-}"

MODULE="${MODULE^^}"
COMMAND="${COMMAND^^}"

case "${MODULE}" in
    SETUP)
        case "${COMMAND}" in
            DATADIR)

                if [ ! -f /data/.datadir_set_up ]; then
                    # if /data is separate partition, probably a Mender-enabled image)
                    # the partition is assumed to be persistent and data is copied
                    if mountpoint /data -q; then
                        cp -r /data_source/. /data
                        echo "OK: (DATADIR) /data_source/ copied to /data/"
                    
                    # otherwise create symlink
                    else
                        if [[ $OVERLAYROOT_ENABLED -eq 1 ]]; then
                            # if overlayroot enabled, create symlink to ssd within overlayroot-chroot, 
                            # will only be ready after reboot
                            mkdir -p /mnt/ssd/data
                            overlayroot-chroot /bin/bash -c "ln -sfn /mnt/ssd/data /"

                            # also create link in tmpfs until next reboot
                            ln -sfn /mnt/ssd/data /
                            echo "OK: (DATADIR) symlink /data --> /mnt/ssd/data created in OVERLAYROOTFS"
                            
                            if [ ! -f /data/.datadir_set_up ]; then
                                cp -r /data_source/* /data
                                echo "OK: (DATADIR) /data_source/ copied to /data/"
                            fi
                            
                        else
                            ln -sfn /data_source /data
                            echo "OK: (DATADIR) symlink /data/ --> /data_source/ created"
                        fi
                    fi
                else
                    echo "WARN: (DATADIR) data directory already set up (found file /data/.datadir_set_up)"
                fi
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
        esac
        ;;

    BITCOIND)
        case "${COMMAND}" in
            RESYNC|REINDEX)
                # stop systemd services
                systemctl stop electrs.service
                systemctl stop lightningd.service
                systemctl stop bitcoind.service

                # deleting bitcoind chainstate in /mnt/ssd/bitcoin/.bitcoin/chainstate
                rm -rf /mnt/ssd/bitcoin/.bitcoin/chainstate
                rm -rf /mnt/ssd/electrs/db

                # for RESYNC incl. download, delete `blocks` directory too
                if [[ "${COMMAND}" == "RESYNC" ]]; then
                    rm -rf /mnt/ssd/bitcoin/.bitcoin/blocks
                fi

                redis_set "bitcoind:ibd" 1
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

            REFRESH_RPCAUTH)
                # called from systemd-bitcoind-startpre.sh
                # make sure rpc credentials update succeeds, otherwise refresh again
                redis_set "bitcoind:refresh-rpcauth" 1

                # generate rpcauth, store values directly in Redis:
                # bitcoind:rpcauth / bitcoind:rpcuser / bitcoind:rpcpassword
                /opt/shift/scripts/bitcoind-rpcauth.py base

                # recreate config files, taking overlayroot into account
                generateConfig "bitcoin.conf.template"
                generateConfig "lightningd.conf.template"
                generateConfig "electrs.conf.template"
                generateConfig "bbbmiddleware.conf.template"
                generateConfig "bashrc-custom.template"

                echo "INFO: created new bitcoind rpc credentials, updated config files"
                echo "Command ${MODULE} ${COMMAND} successfully executed."
                redis_set "bitcoind:refresh-rpcauth" 0
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
        esac
        ;;

    BASE)
        case "${COMMAND}" in
            RESTART)
                ( sleep 5 ; reboot ) & 
                ;;
            SHUTDOWN)
                ( sleep 5 ; shutdown now ) & 
                ;;
            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
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
                    errorExit FLASHDRIVE_CHECK_NONE

                else
                    echo "FLASHDRIVE CHECK: too many targets found (${flashdrive_count} in total)"
                    errorExit FLASHDRIVE_CHECK_MULTI
                fi               
                ;;

            MOUNT)
                # ensure mountpoint is available
                mkdir -p /mnt/backup

                # check if ARG is valid flashdrive
                if ! lsblk "${ARG}" > /dev/null 2>&1; then
                    echo "FLASHDRIVE MOUNT: device ${ARG} not found."
                    errorExit FLASHDRIVE_MOUNT_NOT_FOUND

                elif [ "$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8 "${ARG}" | wc -l)" -ne 1 ]; then
                    echo "FLASHDRIVE MOUNT: device ${ARG} is not unique and/or has partitions."
                    errorExit FLASHDRIVE_MOUNT_NOT_UNIQUE

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
                        errorExit FLASHDRIVE_MOUNT_NOT_SUPPORTED
                    fi

                fi
                ;;

            UNMOUNT)
                if ! mountpoint /mnt/backup -q; then
                    echo "FLASHDRIVE UNMOUNT: no drive mounted at /mnt/backup"
                    errorExit FLASHDRIVE_UNMOUNT_NOT_MOUNTED
                else
                    umount /mnt/backup
                    echo "FLASHDRIVE UNMOUNT: /mnt/backup unmounted."
                fi
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
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
                    errorExit BACKUP_SYSCONFIG_NOT_A_MOUNTPOINT
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
                errorExit CMD_SCRIPT_INVALID_ARG
        esac
        ;;    

    RESTORE)
        REDIS_FILEPATH="/data/redis/bitboxbase.rdb"
        HSM_FILEPATH="/mnt/ssd/bitcoin/.lightning/hsm_secret"

        case "${COMMAND}" in
            # restore system configuration from mounted usb flashdrive
            SYSCONFIG)
                if [ -f /mnt/backup/bbb-backup.rdb ]; then
                    systemctl stop redis.service
                    cp "/mnt/backup/bbb-backup.rdb" "${REDIS_FILEPATH}"
                    chown redis:redis "${REDIS_FILEPATH}"
                    systemctl start redis.service
                else
                    echo "ERR: backup file /mnt/backup/bbb-backup.rdb not found"
                    errorExit RESTORE_SYSCONFIG_BACKUP_NOT_FOUND
                fi
                echo "OK: backup file /mnt/backup/bbb-backup.rdb restored to ${REDIS_FILEPATH}."
                ;;

            # restore c-lightning on-chain keys from Redis database
            HSM_SECRET)
                # create snapshot of 'hsm_secret'
                if [ -f "${HSM_FILEPATH}" ]; then
                    cp -p "${HSM_FILEPATH}" "${HSM_FILEPATH}_$(date '+%Y%m%d-%H%M').backup"
                else
                    echo "WARN: no previous 'hsm_secret' found, no local backup created"
                fi

                # save base64 encoded 'hsm_secret' as binary file to file system
                # redis-cli causes script to terminate when Redis not available
                redis-cli GET lightningd:hsm_secret | base64 -d > /mnt/ssd/bitcoin/.lightning/hsm_secret
                chown bitcoin:bitcoin /mnt/ssd/bitcoin/.lightning/hsm_secret
                echo "OK: backup of file 'hsm_secret' restored"
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
        esac
        ;;    

    MENDER-UPDATE)
        # check if mender application is available
        if ! mender --version 2>/dev/null; then 
            echo "ERR: image is not Mender enabled."
            errorExit MENDER_UPDATE_IMAGE_NOT_MENDER_ENABLED
        fi

        case "${COMMAND}" in
            # initiate Mender update from URL
            INSTALL)
                if [[ ${#} -lt 3 ]]; then
                    echo "ERR: no version number (e.g. 0.0.2) supplied"
                    errorExit MENDER_UPDATE_NO_VERSION
                fi

                # check for valid version number
                regex='^([0-9]+)\.([0-9]+)\.([0-9]+)$'
                if [[ ${ARG} =~ ${regex} ]]; then
                    if mender -install "https://github.com/digitalbitbox/bitbox-base/releases/download/${ARG}/BitBoxBase-v${ARG}-RockPro64.base"; then
                        redis_set "base:updating" 10
                    
                    else
                        # Todo(Stadicus): catch the specific error 'expecting signed artifact, but no signature file found'
                        ERR=${?}
                        echo "ERR: mender install failed with error code ${ERR}"
                        errorExit MENDER_UPDATE_INSTALL_FAILED
                    fi

                else
                    echo "ERR: '${ARG}' is not a valid version number"
                    errorExit MENDER_UPDATE_INVALID_VERSION
                fi
                echo "OK: mender update successfully installed, please restart"
                ;;

            # commit Mender update
            COMMIT)

                if mender -commit; then
                    redis_set "base:updating" 40
                
                else
                    ERR=${?}
                    echo "ERR: mender commit failed with error code ${ERR}"
                    errorExit MENDER_UPDATE_COMMIT_FAILED
                fi        
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
        esac
        ;;

    *)
        echo "Invalid argument: module ${MODULE} unknown."
        errorExit CMD_SCRIPT_INVALID_ARG
esac
