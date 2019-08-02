#!/bin/bash

apt install -y tmux unzip smartmontools
mkdir src && cd $_
wget https://github.com/ncw/stressdisk/releases/download/v1.0.12/stressdisk_1.0.12_linux_arm64.zip
unzip stressdisk_1.0.12_linux_arm64.zip
chmod +x stressdisk
mv stressdisk /usr/sbin

wget https://github.com/digitalbitbox/bitbox-base/releases/download/wip/bbbfancontrol.tar.gz
tar xvf bbbfancontrol.tar.gz
chmod +x bbbfancontrol
mv bbbfancontrol /usr/sbin/



echo "/dev/nvme0n1p1 /mnt/ssd ext4 rw,nosuid,dev,noexec,noatime,nodiratime,auto,nouser,async,nofail 0 2" >> /etc/fstab
mount -a

mkdir -p /mnt/ssd/stressdisk

cat << EOF > run-ssd-stressdisk.sh
#!/bin/sh
tmux new-session -d 'watch smartctl -a /dev/nvme0n1'
tmux split-window -h 'stressdisk cycle /mnt/ssd/stressdisk'
tmux split-window -v 'htop'
tmux split-window -v 'bbbfancontrol -v'
tmux -2 attach-session -d
EOF

source run-ssd-stressdisk.sh