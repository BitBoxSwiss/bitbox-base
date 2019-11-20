#!/bin/bash
# shellcheck disable=SC1091
#
# This script is called by the startup-checks.service on boot
# to check basic system parameters and assure correct configuration.
#
# The Redis config mgmt is not available this early in the boot process.
#

set -eu

# --- generic functions --------------------------------------------------------

# include function: exec_overlayroot()
source /opt/shift/scripts/include/exec_overlayroot.sh.inc

# ------------------------------------------------------------------------------

# SSD configuration
# ------------------------------------------------------------------------------
## check if SSD mount is configured in /etc/fstab (trailing space is essential!)
if ! grep -q '/mnt/ssd ' /etc/fstab ; then

    ## valid partition present?
    if lsblk | grep -q 'nvme0n1p1' && [[ $(lsblk -o NAME,SIZE -abrnp | grep nvme0n1p1 | cut -f 2 -d " ") -gt 400000000000 ]]; then
        exec_overlayroot all-layers "echo '/dev/nvme0n1p1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2' >> /etc/fstab"

    elif lsblk | grep -q 'sda1' && [[ $(lsblk -o NAME,SIZE -abrnp | grep sda1 | cut -f 2 -d " ") -gt 400000000000 ]]; then
        exec_overlayroot all-layers "echo '/dev/sda1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2' >> /etc/fstab"

    else
        ## if no valid partition present, is image configured for autosetup of SSD?

        if ! mountpoint /mnt/ssd -q && [ -f /opt/shift/config/.autosetup-ssd ]; then
            # run ssd autosetup, and disable it afterwards on success
            if /opt/shift/scripts/autosetup-ssd.sh format auto --assume-yes
            then
                echo "INFO: autosetup-ssd.sh successfully executed"
                /opt/shift/scripts/bbb-config.sh disable autosetup_ssd
            else
                echo "ERR: autosetup-ssd.sh failed"
            fi
        fi

        ## check for newly created partition (must be bigger than 400GB to prevent flashdrive mount)
        if lsblk | grep -q 'nvme0n1p1' && [[ $(lsblk -o NAME,SIZE -abrnp | grep nvme0n1p1 | cut -f 2 -d " ") -gt 400000000000 ]]; then
            echo "/dev/nvme0n1p1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
        elif lsblk | grep -q 'sda1' && [[ $(lsblk -o NAME,SIZE -abrnp | grep sda1 | cut -f 2 -d " ") -gt 400000000000 ]]; then
            echo "/dev/sda1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
        else
            echo "ERR: autosetup partition not found"
        fi
    fi

else
    echo "/mnt/ssd is already specified in /etc/fstab, no action required"

fi

sudo mount -a

## abort check if SSD mount is not successful
if ! mountpoint /mnt/ssd -q; then
    echo "Mounting of SSD failed"
    errorExit SSD_NOT_MOUNTED
fi

# Folders & permissions
# ------------------------------------------------------------------------------
## create missing directories & always set correct owner
## access control lists (setfacl) are used to control permissions of newly created files
chown bitcoin:system /mnt/ssd/

## bitcoind data storage
mkdir -p /mnt/ssd/bitcoin/.bitcoin/testnet3
chown -R bitcoin:bitcoin /mnt/ssd/bitcoin/
chmod -R u+rw,g+r,g-w,o-rwx /mnt/ssd/bitcoin/
setfacl -d -m g::rx /mnt/ssd/bitcoin/.bitcoin/
setfacl -d -m o::- /mnt/ssd/bitcoin/.bitcoin/

## lightningd socket
mkdir -p /mnt/ssd/bitcoin/.lightning
chmod u+rw,g+rx,g-w,o-rwx /mnt/ssd/bitcoin/.lightning
chmod 700 /mnt/ssd/bitcoin/.lightning/* || true
chmod 770 /mnt/ssd/bitcoin/.lightning/lightning-rpc || true

mkdir -p /mnt/ssd/bitcoin/.lightning-testnet
chmod u+rw,g+rx,g-w,o-rwx /mnt/ssd/bitcoin/.lightning-testnet
chmod 700 /mnt/ssd/bitcoin/.lightning-testnet/* || true
chmod 770 /mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc || true

## electrs data storage
mkdir -p /mnt/ssd/electrs/
chown -R electrs:bitcoin /mnt/ssd/electrs/
chmod -R u+rw,g+r,g-w,o-rwx /mnt/ssd/electrs/

## system folders
mkdir -p /var/log/redis
chown -R redis:redis /var/log/redis

mkdir -p /mnt/ssd/prometheus
chown -R prometheus:system /mnt/ssd/prometheus/

mkdir -p /mnt/ssd/system/journal/
rm -rf /var/log/journal
ln -sfn /mnt/ssd/system/journal /var/log/journal


# Configuration Management
# ------------------------------------------------------------------------------
## check if /data directory is already set up
if [ ! -f /data/.datadir_set_up ]; then
    /opt/shift/scripts/bbb-cmd.sh setup datadir
fi

## create missing directories & always set correct owner
mkdir -p /data/ssh
mkdir -p /data/ssl
mkdir -p /data/bbbmiddleware
chown -R root:system /data

mkdir -p /data/redis/
chown -R redis:system /data/redis/


# Networking
# ------------------------------------------------------------------------------
# make sure wired interface eth0 is used if present (set metric to 10, wifi will have > 1000)
ifmetric eth0 10

timedatectl set-ntp true

# allow failure, e.g. if not running on Armbian
echo "180" > /sys/class/hwmon/hwmon0/pwm1 || true

# check for TLS certificate and create it if missing
if [ ! -f /data/ssl/nginx-selfsigned.key ]; then
    mkdir -p /data/ssl/
    openssl req -x509 -nodes -newkey rsa:2048 -keyout /data/ssl/nginx-selfsigned.key -out /data/ssl/nginx-selfsigned.crt -subj "/CN=localhost"
fi

# check for SSH host certificate and create it if missing
if [ ! -f /data/ssh/ssh_host_ecdsa_key ]; then
    ssh-keygen -f /data/ssh/ssh_host_ecdsa_key -N '' -t ecdsa
fi


# Swap configuration
# ------------------------------------------------------------------------------
## check if swapfile exists on ssd
if [ ! -f /mnt/ssd/swapfile ]; then
    if mountpoint /mnt/ssd -q; then
        echo "Creating /mnt/ssd/swapfile."
        fallocate --length 2GiB /mnt/ssd/swapfile
        mkswap /mnt/ssd/swapfile
        chmod u+rw,g-rwx,o-rwx /mnt/ssd/swapfile
    else
        echo "ERR: No swapfile found, but SSD not mounted."
    fi
fi

## check if swapfile is configured in /etc/fstab, and add it if necessary
if ! grep -q '/mnt/ssd/swapfile' /etc/fstab ; then
    echo "/mnt/ssd/swapfile swap swap defaults 0 0" >> /etc/fstab
fi

## if overlayroot disabled swapfile on ssd, enable it again
sed -i 's/#overlayroot:swapfile#//g' /etc/fstab

## mount potentially updated /etc/fstab, activate swapfile
mount -a
swapon /mnt/ssd/swapfile || true
