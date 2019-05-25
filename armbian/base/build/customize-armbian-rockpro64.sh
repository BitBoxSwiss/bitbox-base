#!/bin/bash

# Copyright 2019 Shift Cryptosecurity AG, Switzerland.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#      http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# -----------------------------------------------------------------------------

# BitBox Base: build script for Armbian base image
# 
# repository:    https://github.com/digitalbitbox/bitbox-base
# documentation: https://digitalbitbox.github.io/bitbox-base/
# 
# This script is used when building the BitBox Base Armbian image, but can also
# be run on a fresh Armbian install to configure and run various services. 
# This currently includes the Bitcoin Core, c-lightning, electrs, Prometheus, 
# Grafana, NGINX and an mDNS responder to broadcast to the local subnet.
# ------------------------------------------------------------------------------

set -e

function echoArguments {
  echo "
================================================================================
==> $1
================================================================================
CONFIGURATION:
    USER / PASSWORD:    base / ${BASE_ROOTPW}
    HOSTNAME:           ${BASE_HOSTNAME}
    BITCOIN NETWORK:    ${BASE_BITCOIN_NETWORK}
    WIFI SSID / PWD:    ${BASE_WIFI_SSID} ${BASE_WIFI_PW}
    WEB DASHBOARD:      ${BASE_DASHBOARD_WEB_ENABLED}
    HDMI DASHBOARD:     ${BASE_DASHBOARD_HDMI_ENABLED}
    SSH ROOT LOGIN:     ${BASE_SSH_ROOT_LOGIN}
    SSH PASSWORD LOGIN: ${BASE_SSH_PASSWORD_LOGIN}
    AUTOSETUP SSD:      ${BASE_AUTOSETUP_SSD}

================================================================================
BUILD OPTIONS:
    HDMI OUTPUT:      ${BASE_HDMI_BUILD}
================================================================================
"
}
# Load build configuration, set defaults
source /tmp/overlay/build/build.conf || true
source /tmp/overlay/build/build-local.conf || true

BASE_HOSTNAME=${BASE_HOSTNAME:-"bitbox-base"}
BASE_BITCOIN_NETWORK=${BASE_BITCOIN_NETWORK:-"testnet"}
BASE_AUTOSETUP_SSD=${BASE_AUTOSETUP_SSD:-"false"}
BASE_WIFI_SSID=${BASE_WIFI_SSID:-""}
BASE_WIFI_PW=${BASE_WIFI_PW:-""}
BASE_SSH_ROOT_LOGIN=${BASE_SSH_ROOT_LOGIN:-"false"}
BASE_SSH_PASSWORD_LOGIN=${BASE_SSH_PASSWORD_LOGIN:-"false"}
BASE_DASHBOARD_WEB_ENABLED=${BASE_DASHBOARD_WEB_ENABLED:-"false"}
BASE_HDMI_BUILD=${BASE_HDMI_BUILD:-"true"}

# HDMI dashboard only enabled if image is built to support it
if [[ "${BASE_HDMI_BUILD}" != "true" ]]; then 
  BASE_DASHBOARD_HDMI_ENABLED="false"
fi
BASE_DASHBOARD_HDMI_ENABLED=${BASE_DASHBOARD_HDMI_ENABLED:-"false"}

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

# Disable Armbian script on first boot
rm -f /root/.not_logged_in_yet

echoArguments "Starting build process."
echoArguments "Starting build process." > /opt/shift/config/buildargs.log

set -ex

# Prevent interactive prompts
export DEBIAN_FRONTEND=noninteractive
export APT_LISTCHANGES_FRONTEND=none
export LANG=C LC_ALL="en_US.UTF-8"
export HOME=/root


# USERS & LOGIN-----------------------------------------------------------------
# - group 'bitcoin' covers sensitive information
# - group 'system' is used for service users without sensitive privileges
# - user 'root' is locked for login
# - user 'base' has sudo rights and is used for low-level user access
# - user 'hdmi' has minimal access rights

# add groups
addgroup --system bitcoin
addgroup --system system

# Set root password (either from configuration or random) and lock account
BASE_ROOTPW=${BASE_ROOTPW:-$(< /dev/urandom tr -dc A-Z-a-z-0-9 | head -c32)}
echo "root:${BASE_ROOTPW}" | chpasswd
passwd -l root

# add user 'base' to sudo and other groups (with options for non-interactive cmd)
adduser --ingroup system --disabled-password --gecos "" base
usermod -a -G sudo,bitcoin base
echo "base:${BASE_ROOTPW}" | chpasswd

# Add trusted SSH keys for login
mkdir -p /root/.ssh 
mkdir -p /home/base/.ssh
if [ -f /tmp/overlay/build/authorized_keys ]; then
  cp -f /tmp/overlay/build/authorized_keys /root/.ssh/
  cp -f /tmp/overlay/build/authorized_keys /home/base/.ssh/
else
  echo "No SSH keys file found (base/build/authorized_keys), password login only."
fi
chown -R base:bitcoin /home/base/
chmod -R 700 /home/base/.ssh/

# disable password login for SSH (authorized ssh keys only)
if [ ! "$BASE_SSH_PASSWORD_LOGIN" == "true" ]; then
  sed -i '/PASSWORDAUTHENTICATION/Ic\PasswordAuthentication no' /etc/ssh/sshd_config
  sed -i '/CHALLENGERESPONSEAUTHENTICATION/Ic\ChallengeResponseAuthentication no' /etc/ssh/sshd_config
fi

# disable root login via SSH
if [ ! "$BASE_SSH_ROOT_LOGIN" == "true" ]; then
  sed -i '/PERMITROOTLOGIN/Ic\PermitRootLogin no' /etc/ssh/sshd_config
fi

# add service users 
adduser --system --ingroup bitcoin --disabled-login --home /mnt/ssd/bitcoin/      bitcoin
usermod -a -G system bitcoin
adduser --system --ingroup bitcoin --disabled-login --no-create-home              electrs
usermod -a -G system electrs
adduser --system --group          --disabled-login --home /var/run/avahi-daemon   avahi
adduser --system --ingroup system --disabled-login --no-create-home               prometheus
adduser --system --ingroup system --disabled-login --no-create-home               node_exporter
adduser --system hdmi
chsh -s /bin/bash hdmi

# remove bitcoin user home on rootfs (must be on SSD)
# also revoke direct write access for service users to local directory
if ! mountpoint /mnt/ssd -q; then 
  rm -rf /mnt/ssd/bitcoin/
  chmod 700 /mnt/ssd
fi

apt remove -y ntp network-manager
apt purge -y ntp network-manager


# DEPENDENCIES -----------------------------------------------------------------
curl --retry 5 https://deb.torproject.org/torproject.org/A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89.asc | gpg --import
gpg --export A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89 | apt-key add -
if ! grep -q "deb.torproject.org" /etc/apt/sources.list; then 
  echo "deb https://deb.torproject.org/torproject.org stretch main" >> /etc/apt/sources.list
fi
apt update
apt upgrade -y

# development
apt install -y  git tmux qrencode bwm-ng

# build Bitcoin Core
#apt install -y  build-essential libtool autotools-dev automake pkg-config bsdmainutils python3 \
#                libssl-dev libevent-dev libboost-system-dev libboost-filesystem-dev \
#                libboost-chrono-dev libboost-test-dev libboost-thread-dev libzmq3-dev

# build c-lightning
apt install -y  autoconf automake build-essential git libtool libgmp-dev \
                libsqlite3-dev python python3 net-tools zlib1g-dev libsodium-dev

# build electrs
# apt install -y  clang cmake

# networking
apt install -y  openssl tor net-tools fio fail2ban \
                avahi-daemon avahi-discover libnss-mdns \
                avahi-utils avahi-daemon avahi-discover


# STARTUP CHECKS ---------------------------------------------------------------
cat << 'EOF' > /etc/systemd/system/startup-checks.service
[Unit]
Description=BitBox Base startup checks
After=network-online.target
[Service]
ExecStart=/opt/shift/scripts/startup-checks.sh
Type=simple
[Install]
WantedBy=multi-user.target
EOF


# OS CONFIG --------------------------------------------------------------------
# customize MOTD
echo "MOTD_DISABLE='header tips updates armbian-config'" >> /etc/default/armbian-motd
cat << EOF > /etc/update-motd.d/20-shift
#!/bin/bash
. /etc/os-release
. /etc/armbian-release
KERNELID=$(uname -r)
TERM=linux toilet -f standard -F metal "BitBox Base"
printf '\nWelcome to \e[0;91mARMBIAN\x1B[0m %s %s %s %s\n' "$VERSION $IMAGE_TYPE $PRETTY_NAME $KERNELID"
if ! mountpoint /mnt/ssd -q; then printf '\n\e[0;91mMounting of SSD failed.\x1B[0m \n'; fi
echo "Configured for Bitcoin TESTNET"; echo
EOF
chmod 755 /etc/update-motd.d/20-shift

echo "$BASE_HOSTNAME" > /etc/hostname
hostname -F /etc/hostname

# prepare SSD mount point
mkdir -p /mnt/ssd/

# add shortcuts
cat << EOF > /home/base/.bashrc-custom
export LS_OPTIONS='--color=auto'
alias l='ls $LS_OPTIONS -l'
alias ll='ls $LS_OPTIONS -la'

# Bitcoin
alias bcli='bitcoin-cli -conf=/etc/bitcoin/bitcoin.conf'
alias blog='tail -f /mnt/ssd/bitcoin/.bitcoin/testnet3/debug.log'

# Lightning
alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet'
alias llog='sudo journalctl -f -u lightningd'

# Electrum
alias elog='sudo journalctl -n 100 -f -u electrs'

export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
EOF

echo "source /home/base/.bashrc-custom" >> /home/base/.bashrc
source /home/base/.bashrc-custom

cat << 'EOF' >> /etc/services
electrum-rpc    50000/tcp
electrum        50001/tcp
electrum-tls    50002/tcp
bitcoin         8333/tcp
bitcoin-rpc     8332/tcp
lightning       9735/tcp
middleware      8845/tcp
EOF

# retain journal logs between reboots 
ln -sf /mnt/ssd/system/journal/ /var/log/journal

# make bbb scripts executable by sudo
sudo ln /opt/shift/scripts/bbb-config.sh    /usr/local/sbin/bbb-config.sh
sudo ln /opt/shift/scripts/bbb-systemctl.sh /usr/local/sbin/bbb-systemctl.sh


# SYSTEM CONFIGURATION ---------------------------------------------------------
SYSCONFIG_PATH="/opt/shift/sysconfig"
mkdir -p "${SYSCONFIG_PATH}"
echo "BITCOIN_NETWORK=testnet" > "${SYSCONFIG_PATH}/BITCOIN_NETWORK"


# TOR --------------------------------------------------------------------------
cat << EOF > /etc/tor/torrc
HiddenServiceDir /var/lib/tor/hidden_service_bitcoind/    #BITCOIND#
HiddenServiceVersion 3                                    #BITCOIND#
HiddenServicePort 18333 127.0.0.1:18333                   #BITCOIND#

HiddenServiceDir /var/lib/tor/hidden_service_ssh/         #SSH#
HiddenServiceVersion 3                                    #SSH#
HiddenServicePort 22 127.0.0.1:22                         #SSH#

HiddenServiceDir /var/lib/tor/hidden_service_electrum/    #ELECTRUM#
HiddenServiceVersion 3                                    #ELECTRUM#
HiddenServicePort 50002 127.0.0.1:50002                   #ELECTRUM#

HiddenServiceDir /var/lib/tor/lightningd-service_v2/      #LN2#
HiddenServicePort 9375 127.0.0.1:9735                     #LN2#

HiddenServiceDir /var/lib/tor/lightningd-service_v3/      #LN#
HiddenServiceVersion 3                                    #LN#
HiddenServicePort 9375 127.0.0.1:9735                     #LN#

HiddenServiceDir /var/lib/tor/hidden_service_middleware/  #MIDDLEWARE#
HiddenServiceVersion 3                                    #MIDDLEWARE#
HiddenServicePort 9375 127.0.0.1:8845                     #MIDDLEWARE#
EOF


# BITCOIN ----------------------------------------------------------------------
BITCOIN_VERSION="0.18.0"

mkdir -p /usr/local/src/bitcoin
cd /usr/local/src/bitcoin/
curl --retry 5 -SL https://bitcoincore.org/keys/laanwj-releases.asc | gpg --import
curl --retry 5 -SLO https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}/SHA256SUMS.asc
curl --retry 5 -SLO https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}/bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz

gpg --refresh-keys || true
gpg --verify SHA256SUMS.asc || exit 1
grep "bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz\$" SHA256SUMS.asc | sha256sum -c - || exit 1
tar --strip-components 1 -xzf bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz
install -m 0755 -o root -g root -t /usr/bin bin/*

mkdir -p /etc/bitcoin/
cat << EOF > /etc/bitcoin/bitcoin.conf
# network
testnet=1

# server
server=1
listen=1
daemon=1
txindex=0
prune=0
disablewallet=1
pid=/run/bitcoind/bitcoind.pid
rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie

# rpc
rpcconnect=127.0.0.1

# performance
dbcache=2000
maxmempool=50
maxconnections=40
maxuploadtarget=5000

# tor
proxy=127.0.0.1:9050
seednode=nkf5e6b7pl4jfd4a.onion
seednode=xqzfakpeuvrobvpj.onion
seednode=tsyvzsqwa2kkf6b2.onion
EOF

cat << 'EOF' > /etc/systemd/system/bitcoind.service
[Unit]
Description=Bitcoin daemon
After=network-online.target startup-checks.service tor.service
Requires=startup-checks.service
[Service]
# give Tor some time to provide the SOCKS proxy
ExecStartPre=/bin/bash -c "sleep 10"
ExecStart=/usr/bin/bitcoind -daemon -conf=/etc/bitcoin/bitcoin.conf 
# provide cookie authentication as .env file for electrs and base-middleware 
ExecStartPost=/bin/bash -c " \
  sleep 10 && \
  echo -n 'RPCPASSWORD=' > /mnt/ssd/bitcoin/.bitcoin/.cookie.env && \
  tail -c +12 /mnt/ssd/bitcoin/.bitcoin/.cookie >> /mnt/ssd/bitcoin/.bitcoin/.cookie.env && \
  sleep 10"
RuntimeDirectory=bitcoind
User=bitcoin
Group=bitcoin
Type=forking
PIDFile=/run/bitcoind/bitcoind.pid
Restart=always
RestartSec=60
TimeoutSec=300
PrivateTmp=true
ProtectSystem=full
NoNewPrivileges=true
PrivateDevices=true
MemoryDenyWriteExecute=true
[Install]
WantedBy=multi-user.target
EOF


# LIGHTNING --------------------------------------------------------------------
LIGHTNING_VERSION="0.7.0"

cd /usr/local/src/
git clone https://github.com/ElementsProject/lightning.git || true
cd lightning
git checkout v${LIGHTNING_VERSION}
./configure
make
make install

mkdir -p /etc/lightningd/
cat << EOF > /etc/lightningd/lightningd.conf
bitcoin-cli=/usr/bin/bitcoin-cli
bitcoin-rpcconnect=127.0.0.1
bitcoin-rpcport=18332
network=testnet
lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet
bind-addr=127.0.0.1:9735
proxy=127.0.0.1:9050
log-level=debug
daemon
plugin=/opt/shift/scripts/prometheus-lightningd.py
EOF

cat << 'EOF' >/etc/systemd/system/lightningd.service
[Unit]
Description=c-lightning daemon
Wants=bitcoind.service
After=bitcoind.service
[Service]
ExecStartPre=/bin/systemctl is-active bitcoind.service
ExecStart=/usr/local/bin/lightningd --daemon --conf=/etc/lightningd/lightningd.conf
RuntimeDirectory=lightningd
User=bitcoin
Group=bitcoin
Type=forking
#PIDFile=/run/lightningd/lightningd.pid
Restart=always
RestartSec=10
TimeoutSec=240
PrivateTmp=true
ProtectSystem=full
NoNewPrivileges=true
PrivateDevices=true
MemoryDenyWriteExecute=true
[Install]
WantedBy=multi-user.target
EOF


# ELECTRS ----------------------------------------------------------------------
ELECTRS_VERSION="0.6.1"

# cross-compilation from source is currently not possible
# ---
# mkdir -p /usr/local/src/rust
# cd /usr/local/src/rust
# curl https://static.rust-lang.org/dist/rust-1.34.1-aarch64-unknown-linux-gnu.tar.gz -o rust.tar.gz
# if ! echo "0565e50dae58759a3a5287abd61b1a49dfc086c4d6acf2ce604fe1053f704e53 rust.tar.gz" | sha256sum -c -; then
#   echo "sha256sum for rust.tar.gz failed" >&2
#   exit 1
# fi
# tar --strip-components 1 -xzf rust.tar.gz
# ./install.sh
#
# apt install clang cmake
# git clone https://github.com/romanz/electrs || true
# cd electrs
# git checkout tags/v${ELECTRS_VERSION}
# cargo build --release
# cp /usr/local/src/rust/electrs/target/release/electrs /usr/bin/

mkdir -p /usr/local/src/electrs/
cd /usr/local/src/electrs/
# temporary storage of 'electrs' until official binary releases are available
curl --retry 5 -SLO https://github.com/Stadicus/electrs-bin/raw/master/electrs-${ELECTRS_VERSION}-aarch64-linux-gnu.tar.gz
if ! echo "1b1664afe338aa707660bc16b2d82919e5cb8f5fd35faa61c27a7fef24aad156  electrs-0.6.1-aarch64-linux-gnu.tar.gz" | sha256sum -c -; then
  echo "sha256sum for precompiled 'electrs' failed" >&2
  exit 1
fi
tar -xzf electrs-${ELECTRS_VERSION}-aarch64-linux-gnu.tar.gz -C /usr/bin
chmod +x /usr/bin/electrs

mkdir -p /etc/electrs/
cat << EOF > /etc/electrs/electrs.conf
NETWORK=testnet
RPCCONNECT=127.0.0.1
RPCPORT=18332
DB_DIR=/mnt/ssd/electrs/db
VERBOSITY=vvvv
RUST_BACKTRACE=1
EOF

cat << 'EOF' > /etc/systemd/system/electrs.service
[Unit]
Description=Electrs server daemon
Wants=bitcoind.service
After=bitcoind.service
[Service]
EnvironmentFile=/etc/electrs/electrs.conf
EnvironmentFile=/mnt/ssd/bitcoin/.bitcoin/.cookie.env
ExecStartPre=/bin/systemctl is-active bitcoind.service
ExecStart=/bin/bash -c "electrs --network ${NETWORK} -${VERBOSITY} --index-batch-size=10 --jsonrpc-import --db-dir ${DB_DIR} --daemon-rpc-addr ${RPCCONNECT}:${RPCPORT} --cookie __cookie__:${RPCPASSWORD}"
RuntimeDirectory=electrs
User=electrs
Group=bitcoin
Type=simple
KillMode=process
Restart=always
TimeoutSec=120
RestartSec=30
PrivateTmp=true
ProtectSystem=full
NoNewPrivileges=true
PrivateDevices=true
MemoryDenyWriteExecute=true
[Install]
WantedBy=multi-user.target
EOF


# MIDDLEWARE -------------------------------------------------------------------
GO_VERSION="1.12.4"
export GOPATH=/home/base/go

mkdir -p /usr/local/src/go && cd "$_"
curl --retry 5 -SLO https://dl.google.com/go/go${GO_VERSION}.linux-arm64.tar.gz
if ! echo "b7d7b4319b2d86a2ed20cef3b47aa23f0c97612b469178deecd021610f6917df  go1.12.4.linux-arm64.tar.gz" | sha256sum -c -; then exit 1; fi
tar -C /usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz

## bbbfancontrol
## see https://github.com/digitalbitbox/bitbox-base/blob/fan-control/tools/bbbfancontrol/README.md
cd /opt/shift/tools/bbbfancontrol
/usr/local/go/bin/go build -v
cp bbbfancontrol /usr/local/sbin/
cp bbbfancontrol.service /etc/systemd/system/

## base-middleware
mkdir -p "${GOPATH}/src/github.com/shiftdevices" && cd "$_"
git clone https://github.com/shiftdevices/base-middleware || true
cd base-middleware
make native
cp base-middleware /usr/local/sbin/

mkdir -p /etc/base-middleware/
cat << EOF > /etc/base-middleware/base-middleware.conf
BITCOIN_RPCUSER=__cookie__
BITCOIN_RPCPORT=18332
LIGHTNING_RPCPATH=/mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc
EOF

cat << 'EOF' > /etc/systemd/system/base-middleware.service
[Unit]
Description=BitBox Base Middleware
Wants=bitcoind.service lightningd.service electrs.service
After=lightningd.service

[Service]
Type=simple
EnvironmentFile=/etc/base-middleware/base-middleware.conf
EnvironmentFile=/mnt/ssd/bitcoin/.bitcoin/.cookie.env
ExecStart=/usr/local/sbin/base-middleware -rpcuser=${BITCOIN_RPCUSER} -rpcpassword=${RPCPASSWORD} -rpcport=${BITCOIN_RPCPORT} -lightning-rpc-path=${LIGHTNING_RPCPATH}
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# PROMETHEUS -------------------------------------------------------------------
PROMETHEUS_VERSION="2.9.2"
NODE_EXPORTER_VERSION="0.17.0"

## Prometheus
mkdir -p /usr/local/src/prometheus && cd "$_"
curl --retry 5 -SLO https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-arm64.tar.gz
if ! echo "85b85a35bbf413e17bfce2bf86e13bd37a9e2c753745821b4472833dc5a85b52  prometheus-2.9.2.linux-arm64.tar.gz" | sha256sum -c -; then exit 1; fi
tar --strip-components 1 -xzf prometheus-${PROMETHEUS_VERSION}.linux-arm64.tar.gz

mkdir -p /etc/prometheus /var/lib/prometheus
cp prometheus promtool /usr/local/bin/
cp -r consoles/ console_libraries/ /etc/prometheus/
chown -R prometheus /etc/prometheus /var/lib/prometheus

cat << 'EOF' > /etc/prometheus/prometheus.yml
global:
  scrape_interval:     5m
  evaluation_interval: 5m 
scrape_configs:
  - job_name: node
    scrape_interval: 1m
    static_configs:
      - targets: ['127.0.0.1:9100']
  - job_name: base
    scrape_interval: 1m
    static_configs:
      - targets: ['127.0.0.1:8400']
  - job_name: bitcoind
    static_configs:
      - targets: ['127.0.0.1:8334']
  - job_name: electrs
    static_configs:
    - targets: ['127.0.0.1:4224']
  - job_name: lightningd
    static_configs:
    - targets: ['127.0.0.1:9900']    
EOF

cat << 'EOF' > /etc/systemd/system/prometheus.service
[Unit]
Description=Prometheus
After=network-online.target

[Service]
User=prometheus
Group=system
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --web.listen-address="127.0.0.1:9090" \
    --config.file /etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/mnt/ssd/prometheus \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

## Prometheus Node Exporter
curl --retry 5 -SLO https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-arm64.tar.gz
if ! echo "f0d9a8bfed735e93f49a4e8113e96af2dfc90db759164a785b862c704f633569  node_exporter-0.17.0.linux-arm64.tar.gz" | sha256sum -c -; then exit 1; fi
tar --strip-components 1 -xzf node_exporter-${NODE_EXPORTER_VERSION}.linux-arm64.tar.gz
cp node_exporter /usr/local/bin

cat << 'EOF' > /etc/systemd/system/prometheus-node-exporter.service
[Unit]
Description=Prometheus Node Exporter
After=network-online.target

[Service]
User=node_exporter
Group=system
Type=simple
ExecStart=/usr/local/bin/node_exporter
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

## Prometheus Base status exporter
apt install -y python3-pip python3-setuptools
pip3 install wheel
pip3 install prometheus_client

cat << 'EOF' > /etc/systemd/system/prometheus-base.service
[Unit]
Description=Prometheus base exporter
After=network-online.target

[Service]
ExecStart=/opt/shift/scripts/prometheus-base.py
KillMode=process
User=node_exporter
Group=system
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

## Prometheus Bitcoin Core exporter
cat << 'EOF' > /etc/systemd/system/prometheus-bitcoind.service
[Unit]
Description=Prometheus bitcoind exporter
After=network-online.target bitcoind.service

[Service]
ExecStart=/opt/shift/scripts/prometheus-bitcoind.py
KillMode=process
User=bitcoin
Group=bitcoin
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

## Prometheus plugin for c-lightning
pip3 install pylightning
cd /opt/shift/scripts/
curl --retry 5 -SL https://raw.githubusercontent.com/lightningd/plugins/6d0df3c83bd5098ca084b04ba8f589f33a609b8e/prometheus/prometheus.py -o prometheus-lightningd.py
if ! echo "5e020696545e0cd00c2b2b93b49dc9fca55d6c3c56facd685f6098b720230fb3  prometheus-lightningd.py" | sha256sum -c -; then exit 1; fi
chmod +x prometheus-lightningd.py

# GRAFANA ----------------------------------------------------------------------
GRAFANA_VERSION="6.1.4"

mkdir -p /usr/local/src/grafana && cd "$_"
curl --retry 5 -SLO https://dl.grafana.com/oss/release/grafana_${GRAFANA_VERSION}_arm64.deb
if ! echo "47ffae49ee6412b4b04e2b1ac155cab3467c3c0fd437000b1c8948ed7046d331  grafana_6.1.4_arm64.deb" | sha256sum -c -; then exit 1; fi
dpkg -i grafana_${GRAFANA_VERSION}_arm64.deb

cat << 'EOF' >> /etc/grafana/grafana.ini
[server]
http_addr = 127.0.0.1                   #G010#
root_url = http://127.0.0.1:3000/info/  #G011#
[analytics]
reporting_enabled = false               #G020#
check_for_updates = false               #G021#
[users]
allow_sign_up = false                   #G030#
#disable_login_form = true              #G031#
[auth.anonymous]
enabled = true                          #G040#
EOF

cat << 'EOF' > /etc/grafana/provisioning/datasources/prometheus.yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://127.0.0.1:9090
    isDefault: true
    editable: false
EOF

cat << 'EOF' > /etc/grafana/provisioning/dashboards/bitbox-base.yaml
apiVersion: 1
providers:
- name: 'default'
  orgId: 1
  folder: ''
  type: file
  disableDeletion: false
  updateIntervalSeconds: 10 #how often Grafana will scan for changed dashboards
  options:
    path: /opt/shift/config/grafana/dashboard
EOF

mkdir -p /etc/systemd/system/grafana-server.service.d/
cat << 'EOF' > /etc/systemd/system/grafana-server.service.d/override.conf
[Service]
Restart=always
RestartSec=10
PrivateTmp=true
EOF


# NGINX ------------------------------------------------------------------------
apt install -y nginx
rm -f /etc/nginx/sites-enabled/default

cat << 'EOF' > /etc/nginx/nginx.conf
user www-data;
worker_processes 1;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {
  worker_connections 768;
}

stream {
  ssl_certificate /etc/ssl/certs/nginx-selfsigned.crt;
  ssl_certificate_key /etc/ssl/private/nginx-selfsigned.key;
  ssl_session_cache shared:SSL:1m;
  ssl_session_timeout 4h;
  ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
  ssl_prefer_server_ciphers on;

  upstream electrs {
    server 127.0.0.1:50001;
  }
  server {
    listen 50002 ssl;
    proxy_pass electrs;
  }

  upstream electrs_testnet {
    server 127.0.0.1:60001;
  }
  server {
    listen 51002 ssl;
    proxy_pass electrs_testnet;
  }
}

http {
  include /etc/nginx/mime.types;
  default_type application/octet-stream;
  access_log /var/log/nginx/access.log;
  error_log /var/log/nginx/error.log;
  include /etc/nginx/sites-enabled/*.conf;
}
EOF

cat << 'EOF' > /etc/nginx/sites-available/grafana.conf
server {
  listen 80;
  location = / {
    return 301 http://$host/info/d/BitBoxBase/;
  }
  location /info/ {
    proxy_pass http://127.0.0.1:3000/;
  }
}
EOF

if [[ "${BASE_DASHBOARD_WEB_ENABLED}" == "true" ]]; then
  /opt/shift/scripts/bbb-config.sh enable dashboard_web
fi

mkdir -p /etc/systemd/system/nginx.service.d/
cat << 'EOF' > /etc/systemd/system/nginx.service.d/override.conf
[Unit]
After=grafana-server.service startup-checks.service
 
[Service]
Restart=always
RestartSec=10
PrivateTmp=true
EOF

# DASHBOARD OVER HDMI ----------------------------------------------------------
if [[ "${BASE_HDMI_BUILD}" == "true" ]]; then

  sudo apt-get install -y --no-install-recommends xserver-xorg x11-xserver-utils xinit openbox chromium

  cat << 'EOF' > /etc/xdg/openbox/autostart
# Disable any form of screen saver / screen blanking / power management
xset s off
xset s noblank
xset -dpms

# Start Chromium in kiosk mode (fake 'clean exit' to avoid popups)
sed -i 's/"exited_cleanly":false/"exited_cleanly":true/' ~/.config/chromium/'Local State'
sed -i 's/"exited_cleanly":false/"exited_cleanly":true/; s/"exit_type":"[^"]\+"/"exit_type":"Normal"/' ~/.config/chromium/Default/Preferences
chromium --disable-infobars --kiosk --incognito 'http://localhost/info/d/BitBoxBase/bitbox-base?refresh=10s&from=now-24h&to=now&kiosk'
EOF

  # start x-server on user 'hdmi' login
  cat << 'EOF' > /home/hdmi/.bashrc
startx -- -nocursor && exit
EOF

  # enable autologin for user 'hdmi'
  if [[ "${BASE_DASHBOARD_HDMI_ENABLED}" == "true" ]]; then
    /opt/shift/scripts/bbb-config.sh enable dashboard_hdmi
  fi
  
fi

# NETWORK ----------------------------------------------------------------------
cat << 'EOF' > /etc/NetworkManager/NetworkManager.conf
[main]
plugins=ifupdown,keyfile

[ifupdown]
managed=false
EOF

cat << 'EOF' > /etc/systemd/network/ethernet.network
[Match]
Name=eth*

[Network]
DHCP=yes
EOF

cat << 'EOF' > /etc/systemd/resolved.conf
[Resolve]
FallbackDNS=1.1.1.1 8.8.8.8 8.8.4.4 2001:4860:4860::8888 2001:4860:4860::8844
DNSSEC=yes
Cache=yes
EOF

# include Wifi credentials, if specified
if [[ -n "${BASE_WIFI_SSID}" ]]; then
  sed -i '/WPA-SSID/Ic\  wpa-ssid ${BASE_WIFI_SSID}' /opt/shift/config/wifi/wlan0.conf
  sed -i '/WPA-PSK/Ic\  wpa-psk ${BASE_WIFI_PW}' /opt/shift/config/wifi/wlan0.conf
  cp /opt/shift/config/wifi/wlan0.conf /etc/network/interfaces.d/
  echo "WIFI=1" > ${SYSCONFIG_PATH}/WIFI
fi

# mDNS services
sed -i '/PUBLISH-WORKSTATION/Ic\publish-workstation=yes' /etc/avahi/avahi-daemon.conf

cat << EOF > /etc/avahi/services/bitcoind.service
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name>bitcoin</name>
  <service>
    <type>_bitcoin._tcp</type>
    <port>18333</port>
  </service>
</service-group>
EOF

cat << 'EOF' > /etc/avahi/services/electrs.service
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name>bitcoin electrum server</name>
  <service>
    <type>_electrumx-tls._tcp</type>
    <port>50002</port>
  </service>
</service-group>
EOF

cat << 'EOF' > /etc/avahi/services/lightning.service
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name>lightning</name>
  <service>
    <type>_lightning._tcp</type>
    <port>9735</port>
  </service>
</service-group>
EOF

cat << 'EOF' > /etc/avahi/services/bitboxbase.service
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name>bitbox base middleware</name>
  <service>
    <type>_bitboxbase._tcp</type>
    <port>8845</port>
  </service>
</service-group>
EOF

# FINALIZE ---------------------------------------------------------------------

## Clean up
apt-get install -f
apt clean
apt autoremove -y
rm -rf /usr/local/src/*

## Enable services
systemctl daemon-reload
systemctl enable systemd-networkd.service
systemctl enable systemd-resolved.service
systemctl enable systemd-timesyncd.service
systemctl enable bitcoind.service
systemctl enable lightningd.service
systemctl enable electrs.service
systemctl enable bbbfancontrol.service
systemctl enable prometheus.service
systemctl enable prometheus-node-exporter.service
systemctl enable prometheus-base.service
systemctl enable prometheus-bitcoind.service
systemctl enable grafana-server.service
systemctl enable base-middleware.service

# Set to mainnet if configured
if [ "${BASE_BITCOIN_NETWORK}" == "mainnet" ]; then
  /opt/shift/scripts/bbb-config.sh set bitcoin_network mainnet
fi

if [ "${BASE_AUTOSETUP_SSD}" == "true" ]; then
  /opt/shift/scripts/bbb-config.sh enable autosetup_ssd
fi

set +x
echoArguments "Armbian build process finished. Login using SSH Keys or root password."
