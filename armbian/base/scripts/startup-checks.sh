#!/bin/bash
set -e

# check for TLS certificate and create it if missing
if [ ! -f /etc/ssl/private/nginx-selfsigned.key ]; then
  openssl req -x509 -nodes -newkey rsa:2048 -keyout /etc/ssl/private/nginx-selfsigned.key -out /etc/ssl/certs/nginx-selfsigned.crt -subj "/CN=localhost"
fi

timedatectl set-ntp true
echo "255" > /sys/class/hwmon/hwmon0/pwm1

# check if SSD mount is configured in /etc/fstab
if ! grep -q '/mnt/ssd' /etc/fstab ; then
  if lsblk | grep -q 'nvme0n1p1'; then
    echo "/dev/nvme0n1p1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
  elif lsblk | grep -q 'sda1'; then
    echo "/dev/sda1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
  fi
  mount -a
fi

# abort check if SSD mount is not successful
if ! mountpoint /mnt/ssd -q; then 
  echo "Mounting of SSD failed"
  exit 1
fi

# create missing directories & always set correct owner
mkdir -p /mnt/ssd/bitcoin/
chown -R bitcoin:bitcoin /mnt/ssd/bitcoin/
mkdir -p /mnt/ssd/electrs/
chown -R electrs:electrs /mnt/ssd/electrs/
mkdir -p /mnt/ssd/prometheus
chown -R prometheus:prometheus /mnt/ssd/prometheus/
mkdir -p /mnt/ssd/system/journal/

# We set rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie, but there seems to be
# no way to specify where to expect the bitcoin cookie for c-lightning, so let's
# create a symlink at the expected testnet location.
mkdir -p /mnt/ssd/bitcoin/.bitcoin/testnet3/
ln -fs /mnt/ssd/bitcoin/.bitcoin/.cookie /mnt/ssd/bitcoin/.bitcoin/testnet3/.cookie
