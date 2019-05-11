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

  # image configured for autosetup of SSD?
  if ! mountpoint /mnt/ssd -q && [ -f /opt/shift/config/.autosetup_ssd ]; then
    /opt/shift/scripts/autosetup-ssd.sh apply auto --force
    if [ $? -eq 0 ]; then
      rm /opt/shift/config/.autosetup_ssd
    fi
  fi

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
