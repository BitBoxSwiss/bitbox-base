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
    BUILD LIGHTNINGD:   ${BASE_BUILD_LIGHTNINGD}
    HDMI OUTPUT:        ${BASE_HDMI_BUILD}
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
BASE_BUILD_LIGHTNINGD=${BASE_BUILD_LIGHTNINGD:-"true"}

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
# - user 'root' is disabled from logging in with password
# - user 'base' has sudo rights and is used for low-level user access
# - user 'hdmi' has minimal access rights

# add groups
addgroup --system bitcoin
addgroup --system system

# Set root password (either from configuration or random) and lock account
BASE_ROOTPW=${BASE_ROOTPW:-$(< /dev/urandom tr -dc A-Z-a-z-0-9 | head -c32)}
echo "root:${BASE_ROOTPW}" | chpasswd
passwd -l root

# create user 'base' (--gecos "" is used to prevent interactive prompting for user information)
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


# SOFTWARE PACKAGE MGMT --------------------------------------------------------
## update system
apt update
apt upgrade -y

## remove unnecessary packages
apt -y remove git
apt -y remove libllvm* build-essential libtool autotools-dev automake pkg-config gcc gcc-6 libgcc-6-dev \
              alsa-utils* autoconf* bc* bison* bridge-utils* btrfs-tools* bwm-ng* cmake* command-not-found* console-setup* \
              console-setup-linux* crda* dconf-gsettings-backend* dconf-service* debconf-utils* device-tree-compiler* dialog* dirmngr* \
              dnsutils* dosfstools* ethtool* evtest* f2fs-tools* f3* fancontrol* figlet* fio* flex* fping* glib-networking* glib-networking-services* \
              gnome-icon-theme* gnupg2* gsettings-desktop-schemas* gtk-update-icon-cache* haveged* hdparm* hostapd* html2text* ifenslave* iotop* \
              iperf3* iputils-arping* iw* kbd* libatk1.0-0* libcroco3* libcups2* libdbus-glib-1-2* libgdk-pixbuf2.0-0* libglade2-0* libnl-3-dev* \
              libpango-1.0-0* libpolkit-agent-1-0* libpolkit-backend-1-0* libpolkit-gobject-1-0* libpython-stdlib* libpython2.7-stdlib* libssl-dev* \
              man-db* ncurses-term* psmisc* pv* python-avahi* python-pip* python2.7-minimal rsync* screen* shared-mime-info* \
              unattended-upgrades* unicode-data* unzip* vim* wireless-regdb* wireless-tools* wpasupplicant*

# install dependecies
apt install -y --no-install-recommends \
  tmux git openssl network-manager net-tools fio libnss-mdns avahi-daemon avahi-discover avahi-utils fail2ban acl
apt install -y --no-install-recommends ifmetric


# SYSTEM CONFIGURATION ---------------------------------------------------------
SYSCONFIG_PATH="/opt/shift/sysconfig"
mkdir -p "${SYSCONFIG_PATH}"
echo "BITCOIN_NETWORK=testnet" > "${SYSCONFIG_PATH}/BITCOIN_NETWORK"

## store build information
echo "BUILD_DATE='$(date +%Y-%m-%d)'" > "${SYSCONFIG_PATH}/BUILD_DATE"
echo "BUILD_TIME='$(date +%H:%M)'" > "${SYSCONFIG_PATH}/BUILD_TIME"
echo "BUILD_COMMIT='$(cat /opt/shift/config/latest_commit)'" > "${SYSCONFIG_PATH}/BUILD_COMMIT"

## configure swap file (disable Armbian zram, configure custom swapfile on ssd)
sed -i '/ENABLED=/Ic\ENABLED=false' /etc/default/armbian-zram-config
sed -i '/vm.swappiness=/Ic\vm.swappiness=10' /etc/sysctl.conf

## startup checks
cat << 'EOF' > /etc/systemd/system/startup-checks.service
[Unit]
Description=BitBox Base startup checks
After=network-online.target
[Service]
ExecStart=/opt/shift/scripts/systemd-startup-checks.sh
Type=simple
[Install]
WantedBy=multi-user.target
EOF

## customize ssh login message
echo "MOTD_DISABLE='header tips updates armbian-config'" >> /etc/default/armbian-motd
cat << EOF > /etc/update-motd.d/20-shift
#!/bin/bash
source /etc/os-release
source /etc/armbian-release
KERNELID=$(uname -r)
TERM=linux toilet -f standard -F metal "BitBox Base"
printf '\nWelcome to \e[0;91mARMBIAN\x1B[0m %s %s %s %s\n' "$VERSION $IMAGE_TYPE $PRETTY_NAME $KERNELID"
if ! mountpoint /mnt/ssd -q; then printf '\n\e[0;91mMounting of SSD failed.\x1B[0m \n'; fi
echo "Configured for Bitcoin TESTNET"; echo
EOF
chmod 755 /etc/update-motd.d/20-shift

## set hostname
/opt/shift/scripts/bbb-config.sh set hostname "${BASE_HOSTNAME}"

## prepare SSD mount point
mkdir -p /mnt/ssd/

## add shortcuts
cat << EOF > /home/base/.bashrc-custom
export LS_OPTIONS='--color=auto'
alias l='ls $LS_OPTIONS -l'
alias ll='ls $LS_OPTIONS -la'

# Bitcoin
alias bcli='bitcoin-cli -conf=/etc/bitcoin/bitcoin.conf'
alias blog='sudo journalctl -f -u bitcoind'

# Lightning
alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet'
alias llog='sudo journalctl -f -u lightningd'

# Electrum
alias elog='sudo journalctl -n 100 -f -u electrs'
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

## retain journal logs between reboots 
ln -sf /mnt/ssd/system/journal/ /var/log/journal

## make bbb scripts executable with sudo
sudo ln -s /opt/shift/scripts/bbb-config.sh    /usr/local/sbin/bbb-config.sh
sudo ln -s /opt/shift/scripts/bbb-systemctl.sh /usr/local/sbin/bbb-systemctl.sh


# TOR --------------------------------------------------------------------------
curl --retry 5 https://deb.torproject.org/torproject.org/A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89.asc | gpg --import
gpg --export A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89 | apt-key add -
if ! grep -q "deb.torproject.org" /etc/apt/sources.list; then 
  echo "deb https://deb.torproject.org/torproject.org stretch main" >> /etc/apt/sources.list
fi

apt update
apt -y install tor --no-install-recommends

## allow user 'bitcoin' to access Tor proxy socket
usermod -a -G debian-tor bitcoin

cat << EOF > /etc/tor/torrc
ControlPort 9051                                          #TOR#
CookieAuthentication 1                                    #TOR#
CookieAuthFileGroupReadable 1                             #TOR#

HiddenServiceDir /var/lib/tor/hidden_service_ssh/         #SSH#
HiddenServiceVersion 3                                    #SSH#
HiddenServicePort 22 127.0.0.1:22                         #SSH#

HiddenServiceDir /var/lib/tor/hidden_service_electrum/    #ELECTRUM#
HiddenServiceVersion 3                                    #ELECTRUM#
HiddenServicePort 50002 127.0.0.1:50002                   #ELECTRUM#

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
curl --retry 5 -SLO https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}/SHA256SUMS.asc
curl --retry 5 -SLO https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}/bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz

## get Bitcoin Core signing key, verify sha256 checksum of applications and signature of SHA256SUMS.asc
gpg --import /tmp/overlay/build/laanwj-releases.asc
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
listenonion=1
txindex=0
prune=0
disablewallet=1
rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie
sysparms=1
printtoconsole=1

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
ExecStart=/usr/bin/bitcoind -conf=/etc/bitcoin/bitcoin.conf
ExecStartPost=/opt/shift/scripts/systemd-bitcoind-post.sh
RuntimeDirectory=bitcoind
User=bitcoin
Group=bitcoin
Type=simple
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
BIN_DEPS_TAG="v0.0.1-alpha"
LIGHTNING_VERSION="0.7.0"

apt install -y libsodium-dev

## either compile c-lightning from source (default), or use prebuilt binary
if [ "${BASE_BUILD_LIGHTNINGD}" == "true" ]; then
  apt install -y  autoconf automake build-essential git libtool libgmp-dev \
                  libsqlite3-dev python python3 net-tools zlib1g-dev

  cd /usr/local/src/
  git clone https://github.com/ElementsProject/lightning.git
  cd lightning
  git checkout v${LIGHTNING_VERSION}
  ./configure
  make -j 4
  make install

else
  cd /usr/local/src/
  ## temporary storage of 'lightningd' until official arm64 binaries work with stable Armbian release
  curl --retry 5 -SLO https://github.com/digitalbitbox/bitbox-base-deps/releases/download/${BIN_DEPS_TAG}/lightningd_${LIGHTNING_VERSION}-1_arm64.deb
  if ! echo "52be094f8162749acb207bf9ad08125d25288a9d03eb25690f364ba42fcff3c4  lightningd_0.7.0-1_arm64.deb" | sha256sum -c -; then
    echo "sha256sum for precompiled 'lightningd' failed" >&2
    exit 1
  fi
  dpkg -i lightningd_${LIGHTNING_VERSION}-1_arm64.deb

  ## symlink is needed, as the direct compilation (default) installs into /usr/local/bin, while this package uses '/usr/bin'
  ln -s /usr/bin/lightningd /usr/local/bin/lightningd
  
fi

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
plugin=/opt/shift/scripts/prometheus-lightningd.py
EOF

cat << 'EOF' >/etc/systemd/system/lightningd.service
[Unit]
Description=c-lightning daemon
Wants=bitcoind.service
After=bitcoind.service
[Service]
ExecStartPre=/bin/systemctl is-active bitcoind.service
ExecStart=/usr/local/bin/lightningd --conf=/etc/lightningd/lightningd.conf
ExecStartPost=/opt/shift/scripts/systemd-lightningd-post.sh
RuntimeDirectory=lightningd
User=bitcoin
Group=bitcoin
Type=simple
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
BIN_DEPS_TAG="v0.0.2-alpha"
ELECTRS_VERSION="0.7.0"

mkdir -p /usr/local/src/electrs/
cd /usr/local/src/electrs/
## temporary storage of 'electrs' until official binary releases are available
curl --retry 5 -SLO https://github.com/digitalbitbox/bitbox-base-deps/releases/download/${BIN_DEPS_TAG}/electrs-${ELECTRS_VERSION}-aarch64-linux-gnu.tar.gz
if ! echo "77343603d763d5edf31269984551a7aa092afe23127d11b4e6e491522cc029e5  electrs-${ELECTRS_VERSION}-aarch64-linux-gnu.tar.gz" | sha256sum -c -; then
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
DAEMON_DIR=/mnt/ssd/bitcoin/.bitcoin
MONITORING_ADDR=127.0.0.1:4224
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
ExecStart=/usr/bin/electrs \
    --network ${NETWORK} \
    --db-dir ${DB_DIR} \
    --daemon-dir ${DAEMON_DIR} \
    --cookie "__cookie__:${RPCPASSWORD}" \
    --monitoring-addr ${MONITORING_ADDR} \
    -${VERBOSITY}

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

## bbbfancontrol
## see https://github.com/digitalbitbox/bitbox-base/blob/fan-control/tools/bbbfancontrol/README.md
cp /tmp/overlay/bbbfancontrol /usr/local/sbin/
cp /tmp/overlay/bbbfancontrol.service /etc/systemd/system/

## base-middleware
## see https://github.com/digitalbitbox/bitbox-base/blob/master/middleware/README.md
cp /tmp/overlay/base-middleware /usr/local/sbin/

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

  ## start x-server on user 'hdmi' login
  cat << 'EOF' > /home/hdmi/.bashrc
startx -- -nocursor && exit
EOF

  ## enable autologin for user 'hdmi'
  if [[ "${BASE_DASHBOARD_HDMI_ENABLED}" == "true" ]]; then
    /opt/shift/scripts/bbb-config.sh enable dashboard_hdmi
  fi
  
fi

# NETWORK ----------------------------------------------------------------------
cat << 'EOF' > /etc/systemd/resolved.conf
[Resolve]
FallbackDNS=1.1.1.1 8.8.8.8 8.8.4.4 2001:4860:4860::8888 2001:4860:4860::8844
DNSSEC=yes
Cache=yes
EOF

## include Wifi credentials, if specified (experimental)
if [[ -n "${BASE_WIFI_SSID}" ]]; then
  sed -i "/WPA-SSID/Ic\  wpa-ssid ${BASE_WIFI_SSID}" /opt/shift/config/wifi/wlan0.conf
  sed -i "/WPA-PSK/Ic\  wpa-psk ${BASE_WIFI_PW}" /opt/shift/config/wifi/wlan0.conf
  cp /opt/shift/config/wifi/wlan0.conf /etc/network/interfaces.d/
  echo "WIFI=1" > ${SYSCONFIG_PATH}/WIFI
fi

## mDNS services
sed -i '/PUBLISH-WORKSTATION/Ic\publish-workstation=yes' /etc/avahi/avahi-daemon.conf

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

## firewall: restore iptables rules on startup
cat << 'EOF' > /etc/systemd/system/iptables-restore.service
[Unit]
Description=BitBox Base: restore iptables rules
Before=network.target
[Service]
Type=oneshot
ExecStart=/bin/sh -c "/sbin/iptables-restore < /opt/shift/config/iptables/iptables.rules"
[Install]
WantedBy=multi-user.target
EOF


# FINALIZE ---------------------------------------------------------------------

## Remove build-only packages
apt -y remove git

## Delete unnecessary local files
rm -rf /usr/share/doc/*
rm -rf /var/lib/apt/lists/*
rm /usr/bin/test_bitcoin /usr/bin/bitcoin-qt /usr/bin/bitcoin-wallet
find /var/log -maxdepth 1 -type f -delete
locale-gen en_US.UTF-8

## Clean up
apt install -f
apt clean
apt -y autoremove
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
systemctl enable iptables-restore.service

## Set to mainnet if configured
if [ "${BASE_BITCOIN_NETWORK}" == "mainnet" ]; then
  /opt/shift/scripts/bbb-config.sh set bitcoin_network mainnet
fi

if [ "${BASE_AUTOSETUP_SSD}" == "true" ]; then
  /opt/shift/scripts/bbb-config.sh enable autosetup_ssd
fi

set +x
echoArguments "Armbian build process finished. Login using SSH Keys or root password."
