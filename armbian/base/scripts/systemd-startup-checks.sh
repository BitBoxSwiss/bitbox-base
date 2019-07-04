#!/bin/bash
#
# This script is called by the startup-checks.service on boot
# to check basic system parameters and assure correct configuration.
#

set -eux

SYSCONFIG_PATH="/opt/shift/sysconfig"

# check for TLS certificate and create it if missing
if [ ! -f /etc/ssl/private/nginx-selfsigned.key ]; then
  openssl req -x509 -nodes -newkey rsa:2048 -keyout /etc/ssl/private/nginx-selfsigned.key -out /etc/ssl/certs/nginx-selfsigned.crt -subj "/CN=localhost"
fi

# make sure wired interface eth0 is used if present (set metric to 10, wifi will have > 1000)
ifmetric eth0 10

timedatectl set-ntp true
echo "180" > /sys/class/hwmon/hwmon0/pwm1

# disable serial output on UART0
if ! grep -Fiq "console=display" /boot/armbianEnv.txt; then
  if ! grep -Fiq "console=" /boot/armbianEnv.txt; then
    echo "console=display" >> /boot/armbianEnv.txt
  else
    sed -i '/console=/Ic\console=display' /boot/armbianEnv.txt
  fi
  mkimage -C none -A arm -T script -d /boot/boot.cmd /boot/boot.scr > /opt/shift/config/uartconfig.log

  source /opt/shift/sysconfig/UART_REBOOT
  if [ ${UART_REBOOT} -eq 1 ]; then 
    echo "ERR: previous UART_REBOOT not successful, check system"
  else
    echo "UART_REBOOT=1" > /opt/shift/sysconfig/UART_REBOOT
    reboot
  fi

else
  echo "UART_REBOOT=0" > /opt/shift/sysconfig/UART_REBOOT
fi

# check if swapfile exists on ssd
if [ ! -f /mnt/ssd/swapfile ]; then
  fallocate --length 2GiB /mnt/ssd/swapfile
  chmod 600 /mnt/ssd/swapfile
  mkswap /mnt/ssd/swapfile
  swapon /mnt/ssd/swapfile
fi

# check if swapfile is configured in /etc/fstab, and add it if necessary
if ! grep -q '/mnt/ssd/swapfile' /etc/fstab ; then
  echo "/mnt/ssd/swapfile swap swap defaults 0 0" >> /etc/fstab
fi

# check if SSD mount is configured in /etc/fstab
if ! grep -q '/mnt/ssd' /etc/fstab ; then

  # valid partition present?
  if lsblk | grep -q 'nvme0n1p1'; then
    echo "/dev/nvme0n1p1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab

  elif lsblk | grep -q 'sda1'; then
    echo "/dev/sda1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab

  else
    # if no valid partition present, is image configured for autosetup of SSD?
    [ -f "${SYSCONFIG_PATH}/AUTOSETUP_SSD" ] && source "${SYSCONFIG_PATH}/AUTOSETUP_SSD"

    if ! mountpoint /mnt/ssd -q && [[ ${AUTOSETUP_SSD} -eq 1 ]]; then
      /opt/shift/scripts/autosetup-ssd.sh format auto --assume-yes
      if [ $? -eq 0 ]; then
        echo "AUTOSETUP_SSD=0" > "${SYSCONFIG_PATH}/AUTOSETUP_SSD"
      fi
    fi

    # check for newly created partition
    if lsblk | grep -q 'nvme0n1p1'; then
      echo "/dev/nvme0n1p1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
    elif lsblk | grep -q 'sda1'; then
      echo "/dev/sda1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
    else
      echo "ERR: autosetup partition not found"
    fi
  fi

else
  echo "/mnt/ssd is already specified in /etc/fstab, no action required"

fi

sudo mount -a

# abort check if SSD mount is not successful
if ! mountpoint /mnt/ssd -q; then 
  echo "Mounting of SSD failed"
  exit 1
fi

# mount potentially updated /etc/fstab or new swap file
mount -a

# create missing directories & always set correct owner
# access control lists (setfacl) are used to control permissions of newly created files 
chown bitcoin:system /mnt/ssd 

# bitcoin data storage
mkdir -p /mnt/ssd/bitcoin/.bitcoin/testnet3
chown -R bitcoin:bitcoin /mnt/ssd/bitcoin/
setfacl -dR -m g::rx /mnt/ssd/bitcoin/.bitcoin/
setfacl -dR -m o::- /mnt/ssd/bitcoin/.bitcoin/

# electrs data storage
mkdir -p /mnt/ssd/electrs/
chown -R electrs:bitcoin /mnt/ssd/electrs/

# system folders
mkdir -p /mnt/ssd/prometheus
chown -R prometheus:system /mnt/ssd/prometheus/
mkdir -p /mnt/ssd/system/journal/

# set permissions for whole ssd 
# (user:rwx group:r-x other:---)
chmod -R 750 /mnt/ssd

# We set rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie, but there seems to be
# no way to specify where to expect the bitcoin cookie for c-lightning, so let's
# create a symlink at the expected testnet location.
mkdir -p /mnt/ssd/bitcoin/.bitcoin/testnet3/
ln -fs /mnt/ssd/bitcoin/.bitcoin/.cookie /mnt/ssd/bitcoin/.bitcoin/testnet3/.cookie
