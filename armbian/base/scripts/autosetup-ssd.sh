#!/bin/bash
set -eu

# BitBox Base: auto-setup of SSD
#

function usage() {
    printf "
BitBox Base: Auto-Setup SSD

This script checks for potential storage targets and automates the setup process.
Use with caution, data on targets will be deleted.

Usage: $0 <status|format> [device] [--assume-yes]

Examples:
* $0 status                     Get list of potential target devices in system
* $0 format /dev/sda            Set up specified device interactively
* $0 format auto                Detect device and set up interactively
* $0 format auto --assume-yes   WARNING! Execute only in well defined environments!

"
}

function list_targets() {
    echo
    printf "%-16s %-7s %16s    %-30s\n" "DEVICE" "FSTYPE" "SIZE" "OK for storage?"
    printf "%-16s %-7s %16s   %-30s\n" "---------------" "------" "-----------------" "------------------------------"

    device_found=0
    targets_total=0
    blockdev_target_path=""

    # loop over all top-level block devices
    while read blockdev; do
        name=$( echo "${blockdev}" | cut -f 1 -d " " )
        type=$( echo "${blockdev}" | cut -f 2 -d " " )
        size=$( echo "${blockdev}" | cut -f 4 -d " " )
        blockdev_target=""

        # check for existing filesystem
        if [[ ${#type} -gt 0 ]]; then
            blockdev_target="NO: has file system"
        
        # check if at least 200GB
        elif [[ ${size} -lt 200000000000 ]]; then
            blockdev_target="NO: too small (min 200GB)"

        else
            # check top-level device for partitions
            partition_count=0
            while read partition; do
                name_part=$( echo "${partition}" | cut -f 1 -d " " )
                if [[ "${name_part}" =~ "${name}" ]] && [[ "${name_part}" != "${name}" ]]; then
                    partition_count=$((partition_count+1))
                    blockdev_target="NO: has ${partition_count} partition(s)"
                fi
            done <<< "$(lsblk -o NAME,FSTYPE,PARTTYPE,SIZE,TYPE,MAJ:MIN -abrnp -e 1,7,31,179,252)"

            if [[ ${#blockdev_target} -eq 0 ]]; then
                targets_total=$((targets_total + 1))
                blockdev_target="OK"
                blockdev_target_path="${name}"
            fi

            # check if specified device is found
            if [[ "${name}" == "${DEVICE}" ]]; then 
                device_found=1
            fi

        fi

        printf "%-16s %-7s %16s    %-20s\n" "${name}" "${type}" "${size}" "${blockdev_target}"

    # `lsblk` returns only top-level devices (-d) and excludes device-types by major number
    # (see http://www.lanana.org/docs/device-list/devices-2.6+.txt)
    done <<< "$(lsblk -o NAME,FSTYPE,PARTTYPE,SIZE,TYPE,MAJ:MIN -abrnp -e 1,7,31,179,252 -d)"
    
    printf "%-16s %-7s %16s   %-20s\n" "---------------" "------" "-----------------" "------------------------------"
    printf "TOTAL %s potential target blockdevices found\n\n" "${targets_total}"
}


function format() {
    ### DANGER ZONE ###
    (
        echo o # Create a new empty DOS partition table
        echo n # Add a new partition
        echo p # Primary partition
        echo 1 # Partition number
        echo   # First sector (Accept default: 1)
        echo   # Last sector (Accept default: varies)
        echo w # Write changes
    ) | fdisk ${DEVICE}

    case ${DEVICE} in
        # internal drive (e.g. PCIe)
        /dev/nvme*)
            mkfs.ext4 -F "${DEVICE}p1"
            ;;
        # external drive (e.g. USB)
        /dev/sd*)
            mkfs.ext4 -F "${DEVICE}1"
            ;;
    esac
    
    printf "\nDevice ${DEVICE} prepared:\n\n"
    lsblk 
    echo
}


ACTION=${1:-""}
DEVICE=${2:-""}
ASSUMEYES=${3:-""}

if ! [[ "${ACTION}" =~ ^(status|format)$ ]]; then
    usage
    exit 1
fi

if [[ ${UID} -ne 0 ]]; then
    echo "${0}: needs to be run as superuser." >&2
    exit 1
fi

case ${ACTION} in
    status)
        list_targets
        ;;

    format)
        if [ "${#}" -lt 2 ]; then
            usage
            printf "Please specify a DEVICE, e.g. /dev/sda\n\n"
            exit 1
        fi
        
        format_target=0        

        # print and check for potential targets)
        list_targets

        if [[ "${device_found}" -eq 0 ]] && [[ "${DEVICE}" != "auto" ]]; then
            printf "Specified DEVICE '${DEVICE}' not found as potential target.\n\n"
            exit 1
        fi

        # sanity checks ------------------------------
        # auto-detect successful?
        if [[ "${DEVICE}" == "auto" ]]; then
            if [[ ${targets_total} -gt 0 ]]; then
                if [[ ${targets_total} -eq 1 ]]; then 
                    DEVICE="${blockdev_target_path}"
                    printf "Target selected due to AUTO option: ${DEVICE}\n"
                else
                    printf "More than one suitable blockdevice found.\nPlease specify device manually.\n\n"
                    exit 1
                fi
            else
                printf "No suitable blockdevice found.\n\n"
                exit 1
            fi                
        fi
        
        # check for 2 slashes & min length
        device_dashes=$(echo "${DEVICE}" | awk -F"/" '{print NF-1}')
        if [[ device_dashes -lt 2 ]] || [[ ${#DEVICE} -lt 8 ]]; then
            printf "This is not a valid blockdevice: ${DEVICE}\n\n"
            exit 1
        fi

        # assume-yes specified? otherwise ask
        if [[ "${ASSUMEYES}" == "--assume-yes" ]]; then
            format_target=1
        else
            # is device recommended for storage? works only with one recommended drive.
            if [[ "${DEVICE}" != "${blockdev_target_path}" ]]; then
                printf "\nDevice ${DEVICE} is not recommended for storage. Continue?\nType: YES or abort with Ctrl-C\n> "
                read ask_confirmation

                if [[ "${ask_confirmation}" != "YES" ]]; then
                    echo "Aborted."
                    exit 1
                fi
            fi

            printf "\nAre you sure you want to COMPLETELY WIPE device ${DEVICE}?\nContinue?\nType: YES or abort with Ctrl-C\n> "
            read ask_confirmation

            if [[ "${ask_confirmation}" == "YES" ]]; then
                format_target=1
            else
                echo "Aborted."
                exit 1
            fi
        fi

        # partition and format target
        if [[ ${format_target} -eq 1 ]]; then
            format
        fi
        ;;

esac