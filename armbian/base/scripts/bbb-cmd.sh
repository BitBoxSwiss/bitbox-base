#!/bin/bash
# shellcheck disable=SC1091
set -eu

# BitBoxBase: system commands repository
#

# print usage information for script
usage() {
  echo "BitBoxBase: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  setup         <datadir>
  bitcoind      <reindex|resync|refresh_rpcauth>
  flashdrive    <check|mount|unmount>
  backup        <sysconfig|hsm_secret>
  restore       <sysconfig|hsm_secret>
  reset         <auth|config|image|ssd>
  mender-update <install|commit>
  presync       <create|restore>

"
}

# MockMode checks all arguments but does not execute anything
#
# usage: call this script with the ENV variable MOCKMODE set to 1, e.g.
#        $ MOCKMODE=1 ./bbb-cmd.sh
#
MOCKMODE=${MOCKMODE:-0}
checkMockMode() {
    if [[ $MOCKMODE -eq 1 ]]; then
        echo "MOCK MODE enabled"
        echo "OK: ${MODULE} -- ${COMMAND} -- ${ARG}"
        exit 0
    fi
}

# error handling
errorExit() {
    echo "$@" 1>&2
    exit 1
}

# don't load includes for MockMode
if [[ $MOCKMODE -ne 1 ]]; then

    if [[ ! -d /opt/shift/scripts/include/ ]]; then
        echo "ERR: includes directory /opt/shift/scripts/include/ not found, must run on BitBoxBase system. Run in MockMode for testing."
        errorExit SCRIPT_INCLUDES_NOT_FOUND
    fi

    # include function exec_overlayroot(), to execute a command, either within overlayroot-chroot or directly
    source /opt/shift/scripts/include/exec_overlayroot.sh.inc

    # include functions redis_set() and redis_get()
    source /opt/shift/scripts/include/redis.sh.inc

    # include function generateConfig() to generate config files from templates
    source /opt/shift/scripts/include/generateConfig.sh.inc
fi

# ------------------------------------------------------------------------------

# check script arguments
if [[ ${#} -lt 2 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
    usage
    exit 0
elif [[ "${1}" == "-v" ]] || [[ "${1}" == "--version" ]]; then
    echo "bbb-cmd version 0.1"
    exit 0
fi

if [[ $MOCKMODE -ne 1 ]] && [[ ${UID} -ne 0 ]]; then
    echo "${0}: needs to be run as superuser."
    errorExit SCRIPT_NOT_RUN_AS_SUPERUSER
fi

MODULE="${1:-}"
COMMAND="${2:-}"
ARG="${3:-}"
ARG2="${4:-}"

MODULE="$(tr '[:lower:]' '[:upper:]' <<< "${MODULE}")"
COMMAND="$(tr '[:lower:]' '[:upper:]' <<< "${COMMAND}")"

case "${MODULE}" in
    SETUP)
        case "${COMMAND}" in
            DATADIR)
                checkMockMode

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
                            if ! overlayroot-chroot /bin/bash -c "ln -sfn /mnt/ssd/data /"; then
                                echo "ERR: could not create data directory symlink in overlayrootfs"
                            fi

                            # also create link in tmpfs until next reboot
                            ln -sfn /mnt/ssd/data /
                            echo "OK: (DATADIR) symlink /data --> /mnt/ssd/data created in OVERLAYROOTFS"

                        else
                            mkdir -p /data
                            echo "OK: (DATADIR) directory /data/ created"
                        fi

                        if [ ! -f /data/.datadir_set_up ]; then
                            cp -r /data_source/* /data
                            echo "OK: (DATADIR) /data_source/ copied to /data/"
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
                checkMockMode

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
                checkMockMode

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

    FLASHDRIVE)
        # sanitize device name
        ARG="${ARG//[^a-zA-Z0-9_\/]/}"

        case "${COMMAND}" in
            CHECK)
                checkMockMode

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
                checkMockMode

                # ensure mountpoint is available
                mkdir -p /mnt/backup

                # detect flashdrive if no device is provided
                if [[ -z "${ARG}" ]]; then
                    ARG="$(/opt/shift/scripts/bbb-cmd.sh flashdrive check)"
                    echo "INFO: autodetected device ${ARG}"
                fi

                # check if ARG is valid flashdrive
                if ! lsblk "${ARG}" > /dev/null 2>&1; then
                    echo "FLASHDRIVE MOUNT: device ${ARG} not found."
                    errorExit FLASHDRIVE_MOUNT_NOT_FOUND

                elif [ "$(lsblk -o NAME,SIZE,FSTYPE -abrnp -I 8 "${ARG}" | wc -l)" -ne 1 ]; then
                    echo "FLASHDRIVE MOUNT: device ${ARG} is not unique and/or has partitions."
                    errorExit FLASHDRIVE_MOUNT_NOT_UNIQUE

                else
                    if mountpoint /mnt/backup -q; then
                        echo "FLASHDRIVE MOUNT: mountpoint /mnt/backup in use, unmounting..."
                        umount /mnt/backup
                    fi

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
                checkMockMode

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
                checkMockMode

                if mountpoint /mnt/backup -q; then
                    cp "${REDIS_FILEPATH}" "/mnt/backup/bbb-backup.rdb"

                    # create backup history (restore not yet implemented)
                    cp "${REDIS_FILEPATH}" "/mnt/backup/bbb-backup_$(date '+%Y%m%d-%H%M').rdb"
                else
                    echo "ERR: /mnt/backup is not a mountpoint"
                    errorExit BACKUP_SYSCONFIG_NOT_A_MOUNTPOINT
                fi
                echo "OK: backup created as /mnt/backup/bbb-backup.rdb"

                # add maintenance token
                MAINTENANCE_TOKEN="$(< /dev/urandom tr -dc A-Za-z0-9 | head -c64)"
                MAINTENANCE_TOKEN_HASH=$(echo -n "${MAINTENANCE_TOKEN}" | sha256sum | tr -d "[:space:]-")

                # write maintenance token to usb drive, no linebreak allowed
                printf "%s" "${MAINTENANCE_TOKEN}" > /mnt/backup/.maintenance-token

                # append reset token hash for permission check locally
                echo "${MAINTENANCE_TOKEN_HASH}" >> /data/maintenance-token-hashes
                chmod 600 /data/maintenance-token-hashes
                echo "OK: maintenance token created on flashdrive"

                # make sure the Shift factory token is removed on first backup
                sed -i '/factory token/d' /data/maintenance-token-hashes
                ;;

            # backup c-lightning on-chain keys in 'hsm_secret' into Redis database
            HSM_SECRET)
                checkMockMode

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
                checkMockMode

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
                checkMockMode

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
        if ! mender --version 2>/dev/null && [[ $MOCKMODE -ne 1 ]]; then
            echo "ERR: image is not Mender enabled."
            errorExit MENDER_UPDATE_IMAGE_NOT_MENDER_ENABLED
        fi

        case "${COMMAND}" in
            # initiate Mender update
            INSTALL)

                if [[ ${#} -lt 3 ]]; then
                    echo "ERR: no argument or version number (e.g. 'flashdrive' or '0.0.2') supplied"
                    errorExit MENDER_UPDATE_NO_VERSION
                fi

                # parse input: 'flashdrive' or valid semantic version number
                # expected syntax: xx.xx.xx, with optional -RCx suffix, e.g. 0.1.5 or 11.3.12-RC7
                regex='^([0-9]{1,2})+[.]+([0-9]{1,2})+[.]+([0-9]{1,2})(-RC[0-9]{1})?$'

                if [[ ${ARG} == "flashdrive" ]]; then
                    UPDATE_URI='/mnt/backup/update.base'

                    if [[ ! -f "${UPDATE_URI}" ]]; then
                        echo "ERR: update file '${UPDATE_URI}' not found"
                        errorExit MENDER_UPDATE_INVALID_VERSION
                    fi

                elif [[ ${ARG} =~ ${regex} ]]; then
                    UPDATE_URI="https://github.com/digitalbitbox/bitbox-base/releases/download/${ARG}/BitBoxBase-v${ARG}-RockPro64.base"

                else
                    echo "ERR: '${ARG}' is not a valid version number"
                    errorExit MENDER_UPDATE_INVALID_VERSION
                fi

                # install Mender update
                if mender -install "${UPDATE_URI}"; then
                    redis_set "base:updating" 10

                else
                    # Todo(Stadicus): catch the specific error 'expecting signed artifact, but no signature file found'
                    ERR=${?}
                    echo "ERR: mender install failed with error code ${ERR}"
                    errorExit MENDER_UPDATE_INSTALL_FAILED
                fi
                echo "OK: mender update successfully installed, please restart"
                ;;

            # commit Mender update
            COMMIT)
                checkMockMode

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

    RESET)
        # possbile commands
        #   reset auth:     reset authentication for running BitBoxApp setup wizard again
        #   reset config:   [not yet implemented] reset system configuration to factory defaults from original Redis values
        #   reset image:    [not yet implemented] reflash Base image
        #   reset ssd:      [not yet implemented] wipe ssd with all Bitcoin and Lightning data (funds/channels will be lost)

        # must provide argument '--assume-yes' (e.g. bbb-cmd.sh reset auth --assume-yes) for non-interactive usage

        if [[ "${COMMAND}" != "AUTH" ]] && [[ "${COMMAND}" != "CONFIG" ]] && [[ "${COMMAND}" != "IMAGE" ]] && [[ "${COMMAND}" != "SSD" ]]; then
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
        fi

        ARG="$(tr '[:lower:]' '[:upper:]' <<< "${ARG}")"
        if [[ "${ARG}" != "--ASSUME-YES" ]]; then
            printf "\nThis will reset the BitBoxBase with command '%s'. Continue?\nType: YES or abort with Ctrl-C\n> " "${COMMAND}"
            read -r ask_confirmation

            if [[ "${ask_confirmation}" != "YES" ]]; then
                echo "Aborted."
                errorExit CMD_SCRIPT_MANUAL_ABORT
            fi
            echo "INFO: reset manually confirmed"
        else
            echo "INFO: reset confirmed with '--assume-yes'"
        fi

        case "${COMMAND}" in
            # reset authentication for running BitBoxApp setup wizard again
            AUTH)
                redis_set "base:setup" 0
                redis_set "middleware:passwordSetup" 0
                if ! systemctl restart bbbmiddleware.service; then
                    echo "ERR: could not restart bbbmiddleware.service"
                    errorExit SYSTEMD_SERVICESTART_FAILED
                fi
                echo "INFO: bbbmiddleware.service restarted"
                echo "OK: middleware authentication reset, setup wizard can be run again."
                ;;

            CONFIG)
                # stop services
                if redis-cli save; then
                systemctl stop redis.service
                fi

                # delete old data
                rm -rf /data/bbbmiddleware
                rm -rf /data/redis
                rm -rf /data/ssh
                rm -rf /data/ssl
                rm /data/.datadir_set_up

                # re-initialize data dir
                /opt/shift/scripts/bbb-cmd.sh setup datadir

                # start services
                systemctl start redis.service

                # recreate directories and certificates (like on first boot)
                /opt/shift/scripts/systemd-startup-checks.sh
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG
        esac
        ;;

    PRESYNC)
        # check and mount external drive
        if ! lsblk -o NAME,TYPE -abrnp -e 1,7,31,179,252 | grep part | grep -q "${ARG} " || [ -z "${ARG}" ]; then
            echo "ERR: external drive partition not found (specify e.g. '/dev/sda1')"
            errorExit PRESYNC_EXTERNAL_DRIVE_NOT_FOUND
        fi

        mkdir -p /mnt/ext
        if mountpoint /mnt/ext -q; then
            umount /mnt/ext
        fi
        mount "${ARG}" /mnt/ext

        # stop bitcoin services
        systemctl stop bitcoind
        systemctl stop lightningd
        systemctl stop electrs

        case "${COMMAND}" in
            # create snapshot of blockchain data
            # bbb-cmd.sh presync create /dev/sdb1
            CREATE)
                checkMockMode

                if [ ! -d /mnt/ssd/bitcoin/.bitcoin ] || [ ! -d /mnt/ssd/electrs/db/mainnet ]; then
                    echo "ERR: required directories not found (run with --help for additional details)" >&2
                    sleep 5
                    errorExit PRESYNC_DIRECTORIES_NOT_FOUND
                fi

                # freespace=$(df -k /mnt/ssd  | awk '/[0-9]%/{print $(NF-2)}')
                # if [[ ${freespace} -lt 400000000 ]]; then
                #     echo "ERR: not enough disk space, should at least have 400 GB" >&2
                #     errorExit PRESYNC_NOT_ENOUGH_DISKSPACE
                # fi

                tar cvfW /mnt/ext/bbb-presync-ssd-"$(date '+%Y%m%d-%H%M')".tar \
                    -C /mnt/ssd/ \
                    bitcoin/.bitcoin/chainstate \
                    bitcoin/.bitcoin/blocks \
                    bitcoin/.bitcoin/chainstate \
                    --exclude='IDENTITY' \
                    --exclude='LOG*' \
                    --exclude='*.log' \
                    electrs/db/mainnet

                echo
                echo "OK: Presync archive created."
                ls -lh /mnt/ssd/bbb-presync*
                ;;

            # restore snapshot of blockchain data
            # bbb-cmd.sh presync restore /dev/sdb1 /mnt/ext/bbb-presync-ssd-20191126-2248.tar
            RESTORE)
                checkMockMode

                echo "${ARG2}"
                ls -la "${ARG2}"

                if [[ ! -f ${ARG2} ]]; then
                    echo "ERR: file ${ARG2} not found"
                    errorExit PRESYNC_RESTORE_ARCHIVE_NOT_FOUND
                fi

                # TODO(Stadicus)
                # rm -rf /mnt/ssd/bitcoin/.bitcoin/blocks/*
                # rm -rf /mnt/ssd/bitcoin/.bitcoin/chainstate/*
                # rm -rf /mnt/ssd/electrs/db/mainnet

                tar xvf "${ARG2}" -C /mnt/ssd

                echo
                echo "OK: Presync archive restored."
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                errorExit CMD_SCRIPT_INVALID_ARG

            umount /mnt/ext

        esac
        ;;


    *)
        echo "Invalid argument: module ${MODULE} unknown."
        errorExit CMD_SCRIPT_INVALID_ARG
esac
