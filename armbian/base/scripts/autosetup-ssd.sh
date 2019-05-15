#!/bin/bash
set -eu

# BitBox Base: auto-setup of SSD
#

function usage() {
    printf "\n
BitBox Base: Auto-Setup SSD

This script checks for potential storage targets and automates the setup process.
Use with caution, data on targets will be deleted.

Usage: $0 <status|apply> [device] [--force]

Examples:
* $0 status                 Get list of potential target devices in system
* $0 apply /dev/sda         Set up specified device interactively
* $0 apply auto             Detect device and set up interactively
* $0 apply auto --force     WARNING! Execute only in well defined environments!

"
}

function status() {
    echo
    printf "%-16s %-7s %16s    %-30s\n" "DEVICE" "FSTYPE" "SIZE" "OK for storage?"
    printf "%-16s %-7s %16s   %-30s\n" "---------------" "------" "-----------------" "------------------------------"

    targets_total=0
    blockdev_target_path=""
    while read blockdev; do
        name=$( echo "$blockdev" | cut -f 1 -d " " )
        type=$( echo "$blockdev" | cut -f 2 -d " " )
        size=$( echo "$blockdev" | cut -f 4 -d " " )
        blockdev_target=""

        # check for existing filesystem
        if [[ ${#type} -gt 0 ]]; then
            blockdev_target="NO: has file system"
        
        # check if at least 200GB
        elif [[ ${size} -lt 200000000000 ]]; then
            blockdev_target="NO: too small (min 200GB)"
        
        else
            # check for sub-partitions
            partition_count=0
            while read partition; do
                name_part=$( echo "$partition" | cut -f 1 -d " " )
                if [[ "$name_part" =~ "$name" ]] && [[ "$name_part" != "$name" ]]; then
                    partition_count=$((partition_count+1))
                    blockdev_target="NO: has $partition_count partition(s)"
                fi
            done <<< "$(lsblk -o NAME,FSTYPE,PARTTYPE,SIZE,TYPE,MAJ:MIN -abrnp -e 1,7,31,179,252)"

            if [[ ${#blockdev_target} -eq 0 ]]; then
                targets_total=$((targets_total + 1))
                blockdev_target="OK"
                blockdev_target_path="$name"
            fi
        fi

        printf "%-16s %-7s %16s    %-20s\n" "$name" "$type" "$size" "$blockdev_target"

    done <<< "$(lsblk -o NAME,FSTYPE,PARTTYPE,SIZE,TYPE,MAJ:MIN -abrnp -e 1,7,31,179,252 -d)"
    printf "%-16s %-7s %16s   %-20s\n" "---------------" "------" "-----------------" "------------------------------"
    printf "TOTAL %s potential target blockdevices found\n" "$targets_total"
}


ACTION=${1:-"help"}
DEVICE=${2:-""}
FORCE=${3:-""}

if ! [[ ${ACTION} =~ ^(status|apply)$ ]]; then
    usage
    exit 1
fi

if [[ ${UID} -ne 0 ]]; then
    echo "${0}: needs to be run as superuser." >&2
    exit 1
fi

case ${ACTION} in
    status)
        status
        ;;

    apply)
        if [ "$#" -lt 2 ]; then
            usage
            printf "Please specify a DEVICE, e.g. /dev/sda\n\n"
            exit 1
        fi
        
        status
        doit=0

        # sanity checks ------------------------------
        # auto-detect successful?
        if [[ "$DEVICE" == "auto" ]]; then
            if [[ $targets_total -gt 0 ]]; then
                if [[ $targets_total -eq 1 ]]; then 
                    DEVICE="$blockdev_target_path"
                    printf "Target selected due to AUTO option: $DEVICE\n"
                else
                    printf "More than one suitable blockdevice found.\nPlease specify device manually.\n\n"
                    exit 1
                fi
            else
                printf "No suitable blockdevice found.\n\n"
                exit 1
            fi                
        fi
        
        # check for 2 dashes & min length
        device_dashes=$(echo "$DEVICE" | awk -F"/" '{print NF-1}')
        if [[ device_dashes -lt 2 ]] || [[ ${#DEVICE} -lt 8 ]]; then
            printf "This is not a valid blockdevice: $DEVICE\n\n"
            exit 1
        fi

        # force specified? otherwise ask
        if [[ "$FORCE" == "--force" ]]; then
            doit=1
        else
            # is device recommended for storage? works only with one recommended drive.
            if [[ "$DEVICE" != "$blockdev_target_path" ]]; then
                printf "\nDevice $DEVICE is not recommended for storage. Continue?\nType: YES or abort with Ctrl-C\n> "
                read type_yes

                if [[ "$type_yes" != "YES" ]]; then
                    echo "Aborted."
                    exit 1
                fi
            fi

            printf "\nAre you sure you want to COMPLETELY WIPE device $DEVICE?\nContinue?\nType: YES or abort with Ctrl-C\n> "
            read type_yes

            if [[ "$type_yes" == "YES" ]]; then
                doit=1
            else
                echo "Aborted."
                exit 1
            fi
        fi

        if [[ doit -eq 1 ]]; then
            ### DANGER ZONE ###
            #parted --script $DEVICE mklabel gpt
            (
                echo o # Create a new empty DOS partition table
                echo n # Add a new partition
                echo p # Primary partition
                echo 1 # Partition number
                echo   # First sector (Accept default: 1)
                echo   # Last sector (Accept default: varies)
                echo w # Write changes
            ) | fdisk $DEVICE

            case $DEVICE in
                /dev/nvme*)
                    echo "${DEVICE}p1"
                    ;;
                /dev/sd*)
                    echo "${DEVICE}1"
                ;;
            esac
            
            echo 
            echo "Device $DEVICE prepared:"
            echo 
            lsblk 
            echo
        fi
        ;;

esac