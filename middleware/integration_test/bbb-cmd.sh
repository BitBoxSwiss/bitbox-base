#!/bin/bash
set -eu

# MOCK DEV SCRIPT
# always return without doing anything

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

# ------------------------------------------------------------------------------

# check script arguments
if [[ ${#} == 0 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
    usage
    exit 0
elif [[ "${1}" == "-v" ]] || [[ "${1}" == "--version" ]]; then
    echo "bbb-cmd version 0.1"
    exit 0
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
                echo "OK: ${MODULE} -- ${COMMAND}"
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;

    BITCOIND)
        case "${COMMAND}" in
            RESYNC|REINDEX)
                echo "OK: ${MODULE} -- ${COMMAND}"
                ;;
            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;
        

    BASE)
        case "${COMMAND}" in
            RESTART|SHUTDOWN)
                echo "OK: ${MODULE} -- ${COMMAND}"
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
                echo "/dev/sdb1"
                ;;

            MOUNT|UNMOUNT)
                echo "OK: ${MODULE} -- ${COMMAND}"
                ;;

            *)
                echo "Invalid argument for module ${MODULE}: command ${COMMAND} unknown."
                exit 1
        esac
        ;;        
    
    BACKUP|RESTORE)
        case "${COMMAND}" in
            # backup system configuration to mounted usb flashdrive
            SYSCONFIG|HSM_SECRET)
                echo "OK: ${MODULE} -- ${COMMAND}"
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
