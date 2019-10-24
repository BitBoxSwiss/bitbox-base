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

echoArguments() {
  echo "
================================================================================
==> $1
================================================================================
CONFIGURATION:
    USER / PASSWORD:    base / ${BASE_LOGINPW}
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
    BUILD MODE:         ${BASE_BUILDMODE}
    LINUX DISTRIBUTION: ${BASE_DISTRIBUTION}
    MINIMAL IMAGE:      ${BASE_MINIMAL}
    OVERLAYROOT:        ${BASE_OVERLAYROOT}
    BUILD LIGHTNINGD:   ${BASE_BUILD_LIGHTNINGD}
    HDMI OUTPUT:        ${BASE_HDMI_BUILD}
================================================================================
"
}

importFile() {
  # copies a single file from the repository directory to the root filesystem
  # this makes every file inclusion explicit
  #
  # argument is full rootfs path, with leading slash /
  #
  local REPO_ROOTFS="/opt/shift/rootfs"

  if [ ${#} -eq 0 ] || [ ${#} -gt 1 ]; then
    echo "ERR: importFile() expects exactly one argument"
    exit 1
  fi

  # create directory
  local DIR
  DIR=$(dirname "${1}")
  mkdir -p "${DIR}"

  # strip leading slash and import file
  local FILE="${1#/}"
  if [ -f "${REPO_ROOTFS}/${FILE}" ]; then
    echo "importFile() copying ${FILE}"
    cd "${REPO_ROOTFS}"
    cp -f --parents "${FILE}" /
    cd -
  else
    echo "ERR: generateConfig() template file ${REPO_ROOTFS}/${FILE} not found"
    exit 1
  fi
}

generateConfig() {
  # generates a config file using custom bbbconfgen
  # https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbconfgen
  #
  # argument is template filename, without path
  #
  local TEMPLATES_DIR="/opt/shift/config/templates"

  if [ ${#} -eq 0 ] || [ ${#} -gt 1 ]; then
    echo "ERR: generateConfig() expects exactly one argument"
    exit 1
  fi

  local FILE="${TEMPLATES_DIR}/${1}"
  if [ -f "${FILE}" ]; then
    echo "generateConfig() from ${FILE}"
    /usr/local/sbin/bbbconfgen --template "${FILE}"
  else
    echo "ERR: generateConfig() template file ${FILE} not found"
    exit 1
  fi
}

# get Linux distribution and version
# (works explicitly only on Armbian Debian Stretch, Buster and Ubuntu Bionic)
cat /etc/os-release
# shellcheck disable=SC1091
source /etc/os-release
BASE_DISTRIBUTION=${VERSION_CODENAME}

# Load build configuration, set defaults
# shellcheck disable=SC1091
source /opt/shift/build.conf || true
# shellcheck disable=SC1091
source /opt/shift/build-local.conf || true

BASE_BUILDMODE=${1:-"armbian-build"}
BASE_DISTRIBUTION=${BASE_DISTRIBUTION:-"bionic"}
BASE_MINIMAL=${BASE_MINIMAL:-"true"}
BASE_HOSTNAME=${BASE_HOSTNAME:-"bitbox-base"}
BASE_BITCOIN_NETWORK=${BASE_BITCOIN_NETWORK:-"mainnet"}
BASE_AUTOSETUP_SSD=${BASE_AUTOSETUP_SSD:-"false"}
BASE_WIFI_SSID=${BASE_WIFI_SSID:-""}
BASE_WIFI_PW=${BASE_WIFI_PW:-""}
BASE_SSH_ROOT_LOGIN=${BASE_SSH_ROOT_LOGIN:-"false"}
BASE_SSH_PASSWORD_LOGIN=${BASE_SSH_PASSWORD_LOGIN:-"false"}
BASE_DASHBOARD_WEB_ENABLED=${BASE_DASHBOARD_WEB_ENABLED:-"false"}
BASE_HDMI_BUILD=${BASE_HDMI_BUILD:-"false"}
BASE_BUILD_LIGHTNINGD=${BASE_BUILD_LIGHTNINGD:-"true"}
BASE_OVERLAYROOT=${BASE_OVERLAYROOT:-"false"}

# HDMI dashboard only enabled if image is built to support it
if [[ "${BASE_HDMI_BUILD}" != "true" ]]; then
  echo "WARN: HDMI dashboard is disabled. It cannot be enabled without BASE_HDMI_BUILD option set to 'true'."
  BASE_DASHBOARD_HDMI_ENABLED="false"
fi
BASE_DASHBOARD_HDMI_ENABLED=${BASE_DASHBOARD_HDMI_ENABLED:-"false"}

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

# configuration checks
if [[ "${BASE_DISTRIBUTION}" =~ ^(bionic|buster)$ ]] && [[ "${BASE_BUILD_LIGHTNINGD}" != "true" ]]; then
  echo "ERR: precomplied binaries for c-lightning are not compatible with Debian Buster or Ubuntu Bionic at the moment,"
  echo "     please use the option BASE_BUILD_LIGHTNINGD='true' in build.conf"
  exit 1
fi

# Disable Armbian script on first boot
rm -f /root/.not_logged_in_yet

mkdir -p /opt/shift/config/
echoArguments "Starting build process."
echoArguments "Starting build process." > /opt/shift/config/buildargs.log

set -ex

# Prevent interactive prompts
export DEBIAN_FRONTEND="noninteractive"
export APT_LISTCHANGES_FRONTEND="none"
export LANG=C LC_ALL="en_US.UTF-8"
export HOME=/root


# USERS & LOGIN-----------------------------------------------------------------
# - group 'bitcoin' covers sensitive information
# - group 'system' is used for service users without sensitive privileges
# - user 'root' is disabled from logging in with password
# - user 'base' has sudo rights and is used for low-level user access
# - user 'hdmi' has minimal access rights
# - other users are setup as system user, with disabled login

# add groups
addgroup --system bitcoin
addgroup --system system

# Set root password (either from configuration or random) and lock account
BASE_LOGINPW=${BASE_LOGINPW:-$(< /dev/urandom tr -dc A-Z-a-z-0-9 | head -c32)}
echo "root:${BASE_LOGINPW}" | chpasswd

# create user 'base' (--gecos "" is used to prevent interactive prompting for user information)
adduser --ingroup system --disabled-password --gecos "" base || true
usermod -a -G sudo,bitcoin base
echo "base:${BASE_LOGINPW}" | chpasswd

# Add trusted SSH keys for login
mkdir -p /home/base/.ssh/
if [ -f /opt/shift/config/ssh/authorized_keys ]; then
  cp /opt/shift/config/ssh/authorized_keys /home/base/.ssh/
else
  echo "No SSH keys file found (base/authorized_keys), password login only."
fi
chown -R base:bitcoin /home/base/
chmod -R u+rw,g-rwx,o-rwx /home/base/.ssh

# add service users
adduser --system --ingroup bitcoin --disabled-login --home /mnt/ssd/bitcoin/      bitcoin || true
usermod -a -G system bitcoin

adduser --system --ingroup bitcoin --disabled-login --no-create-home              electrs || true
usermod -a -G system electrs

adduser --system --group          --disabled-login --home /var/run/avahi-daemon   avahi || true

adduser --system --ingroup system --disabled-login --no-create-home               prometheus || true

adduser --system --ingroup system --disabled-login --no-create-home               node_exporter || true

adduser --system --group          --disabled-login --no-create-home               redis || true
usermod -a -G system redis

adduser --system hdmi
chsh -s /bin/bash hdmi

# remove bitcoin user home on rootfs (must be on SSD)
# also revoke direct write access for service users to local directory
if ! mountpoint /mnt/ssd -q; then
  rm -rf /mnt/ssd/bitcoin/
  chmod u+rwx,g-rwx,o-rwx /mnt/ssd
fi


# SOFTWARE PACKAGE MGMT --------------------------------------------------------
## update system, force non-interactive commands

apt -y update
apt -y -q -o "Dpkg::Options::=--force-confdef" -o "Dpkg::Options::=--force-confold" upgrade
apt -y --fix-broken install

## remove unnecessary packages (only when building image, not ondevice)
if [[ "${BASE_BUILDMODE}" != "ondevice" ]] && [[ "${BASE_MINIMAL}" == "true" ]]; then
  pkgToRemove="git libllvmkk build-essential libtool autotools-dev automake pkg-config gcc gcc-6 libgcc-6-dev
  alsa-utils* autoconf* bc* bison* bridge-utils* btrfs-tools* bwm-ng* cmake* command-not-found* console-setup*
  console-setup-linux* crda* dconf-gsettings-backend* dconf-service* debconf-utils* device-tree-compiler* dialog* dirmngr*
  dnsutils* dosfstools* ethtool* evtest* f2fs-tools* f3* fancontrol* figlet* fio* flex* fping* glib-networking* glib-networking-services*
  gnome-icon-theme* gnupg2* gsettings-desktop-schemas* gtk-update-icon-cache* haveged* hdparm* hostapd* html2text* ifenslave* iotop*
  iperf3* iputils-arping* iw* kbd* libatk1.0-0* libcroco3* libcups2* libdbus-glib-1-2* libgdk-pixbuf2.0-0* libglade2-0* libnl-3-dev*
  libpango-1.0-0* libpolkit-agent-1-0* libpolkit-backend-1-0* libpolkit-gobject-1-0* libpython-stdlib* libpython2.7-stdlib* libssl-dev*
  man-db* ncurses-term* psmisc* pv* python-avahi* python-pip* python2.7-minimal screen* shared-mime-info*
  unattended-upgrades* unicode-data* unzip* vim* wireless-regdb* wireless-tools* wpasupplicant* "

  for pkg in $pkgToRemove
  do
    apt -y remove "$pkg" || true
  done

  apt -y --fix-broken install
fi

## install required software packages
apt install -y --no-install-recommends \
  git openssl network-manager net-tools fio libnss-mdns avahi-daemon avahi-discover avahi-utils fail2ban acl rsync smartmontools curl
apt install -y --no-install-recommends ifmetric
apt install -y iptables-persistent

## install python dependencies
apt install -y python3-pip python3-setuptools
pip3 install wheel
pip3 install prometheus_client
pip3 install redis
pip3 install pylightning

# debug
apt install -y --no-install-recommends tmux unzip

if [[ "${BASE_DISTRIBUTION}" == "bionic" ]]; then
    apt install -y --no-install-recommends overlayroot
fi


# REDIS & CONFIGURATION MGMT ---------------------------------------------------

## create data directory
## standard build links from /data to /data_source on first boot, but
## Mender build mounts /data as own partition, data needs to be copied on first boot
##
## create symlink for all scripts to work, remove it at the end of build process
mkdir -p /data_source/
ln -sfn /data_source /data
touch /data/.datadir_set_up

## install Redis
apt install -y --no-install-recommends redis
mkdir -p /data/redis/
chown -R redis:redis /data/redis/

### disable standard systemd unit
systemctl disable redis-server.service
systemctl mask redis-server.service

echo "include /etc/redis/redis-local.conf" >> /etc/redis/redis.conf
importFile /etc/redis/redis-local.conf

importFile /etc/systemd/system/redis.service
systemctl enable redis.service

## start temporary Redis server for build process
redis-server --daemonize yes --databases 1 --dbfilename bitboxbase.rdb --dir /data/redis/

## bulk import factory settings / store build info
< /opt/shift/config/redis/factorysettings.txt sh /opt/shift/scripts/redis-pipe.sh | redis-cli --pipe
redis-cli SET build:date "$(date +%Y-%m-%d)"
redis-cli SET build:time "$(date +%H:%M)"
redis-cli SET build:commit "$(cat /opt/shift/config/latest_commit)"
redis-cli KEYS "*"

## bbbconfgen
## see https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbconfgen
if [ -f /opt/shift/bin/go/bbbconfgen ]; then
  cp /opt/shift/bin/go/bbbconfgen /usr/local/sbin/
else
  echo "INFO: bbbconfgen not found, executable is downloaded from GitHub."
  mkdir -p /usr/local/src/bbbconfgen && cd "$_"
  curl --retry 5 -SLO "https://github.com/digitalbitbox/bitbox-base-deps/releases/download/${BIN_DEPS_TAG}/bbbconfgen.tar.gz"
  #TODO(Stadicus): add checksum once releases are stable
  # if ! echo "xxxxx  bbbconfgen.tar.gz" | sha256sum -c -; then
  #   echo "sha256sum for precompiled 'bbbconfgen' failed" >&2
  #   exit 1
  # fi
  tar -xzf bbbconfgen.tar.gz
  cp ./bbbconfgen /usr/local/sbin/
fi


# SYSTEM CONFIGURATION ---------------------------------------------------------
## create custom default systemd target
## this allows to start custom applications after regular system boot
importFile /etc/systemd/system/bitboxbase.target
ln -sf /etc/systemd/system/bitboxbase.target /etc/systemd/system/default.target

## configure sshd (authorized ssh keys only)
mv /etc/ssh/sshd_config /etc/ssh/sshd_config.original
generateConfig sshd_config.template # --> /etc/ssh/sshd_config
rm -f /etc/ssh/ssh_host_*

## optionally, enable ssh password login
if [ "$BASE_SSH_PASSWORD_LOGIN" == "true" ]; then
  /opt/shift/scripts/bbb-config.sh enable pwlogin
fi

## set hostname
/opt/shift/scripts/bbb-config.sh set hostname "${BASE_HOSTNAME}"

## set debug console to only use display, not serial console ttyS2 over UART
echo 'console=display' >> /boot/armbianEnv.txt
systemctl mask serial-getty@ttyS2.service || true
systemctl mask serial-getty@ttyFIQ0       || true

## generate selfsigned NGINX key when run script is run on device, plus symlink to /data
mkdir -p /data/ssl/
if [ ! -f /data/ssl/nginx-selfsigned.key ] && [[ "${BASE_BUILDMODE}" == "ondevice" ]]; then
  openssl req -x509 -nodes -newkey rsa:2048 -keyout /data/ssl/nginx-selfsigned.key -out /data/ssl/nginx-selfsigned.crt -subj "/CN=localhost"
fi

## disable Armbian ramlog and limit logsize if overlayroot is enabled
if [ "$BASE_OVERLAYROOT" == "true" ]; then
  sed -i '/ENABLED=/Ic\ENABLED=false' /etc/default/armbian-ramlog
  sed -i 's/log.hdd/log/g' /etc/logrotate.conf
  importFile /etc/logrotate.d/rsyslog
fi

## retain less NGINX logs
sed -i 's/daily/size 1M/g' /etc/logrotate.d/nginx || true
sed -i '/\trotate/Ic\\trotate 14' /etc/logrotate.d/nginx || true

## configure systemd journal
importFile "/etc/systemd/journald.conf"

## run logroate every 10 minutes
importFile "/etc/systemd/system/logrotate.service"
importFile "/etc/systemd/system/logrotate.timer"
systemctl enable logrotate.timer

## retain journal logs between reboots on the SSD
rm -rf /var/log/journal
ln -sfn /mnt/ssd/system/journal /var/log/journal

## configure mender artifact verification key
importFile "/etc/mender/mender.conf"

## configure swap file (disable Armbian zram, configure custom swapfile on ssd)
sed -i '/ENABLED=/Ic\ENABLED=false' /etc/default/armbian-zram-config
sed -i '/vm.swappiness=/Ic\vm.swappiness=10' /etc/sysctl.conf

## startup checks
importFile /etc/systemd/system/startup-checks.service
systemctl enable startup-checks.service

## update checks
importFile /etc/systemd/system/update-checks.service
systemctl enable update-checks.service

## disable ssh login messages
echo "MOTD_DISABLE='header tips updates armbian-config'" >> /etc/default/armbian-motd

## prepare SSD mount point
mkdir -p /mnt/ssd/

## add bash shortcuts
generateConfig bashrc-custom.template # -->  /home/base/.bashrc-custom
chown base:bitcoin /home/base/.bashrc-custom
chmod u+rw,g-rwx,o-rwx /home/base/.bashrc-custom
echo "source /home/base/.bashrc-custom" >> /home/base/.bashrc
# shellcheck disable=SC1091
source /home/base/.bashrc-custom

cat << 'EOF' >> /etc/services
electrum-rpc    50000/tcp
electrum        50001/tcp
electrum-tls    50002/tcp
bitcoin         8333/tcp
bitcoin-rpc     8332/tcp
lightning       9735/tcp
bbbmiddleware   8845/tcp
EOF

## make bbb scripts executable with sudo
ln -sf /opt/shift/scripts/bbb-config.sh    /usr/local/sbin/bbb-config.sh
ln -sf /opt/shift/scripts/bbb-cmd.sh       /usr/local/sbin/bbb-cmd.sh
ln -sf /opt/shift/scripts/bbb-systemctl.sh /usr/local/sbin/bbb-systemctl.sh


# TOR --------------------------------------------------------------------------
curl --retry 5 https://deb.torproject.org/torproject.org/A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89.asc | gpg --import
gpg --export A3C4F0F979CAA22CDBA8F512EE8CBC9E886DDD89 | apt-key add -
if ! grep -q "deb.torproject.org" /etc/apt/sources.list; then
  echo "deb https://deb.torproject.org/torproject.org ${BASE_DISTRIBUTION} main" >> /etc/apt/sources.list
fi

apt update
apt -y install tor --no-install-recommends
generateConfig "torrc.template" # --> /etc/tor/torrc

## allow user 'bitcoin' to access Tor proxy socket
usermod -a -G debian-tor bitcoin



# BITCOIN ----------------------------------------------------------------------
BITCOIN_VERSION="0.18.1"

mkdir -p /usr/local/src/bitcoin
cd /usr/local/src/bitcoin/
curl --retry 5 -SLO "https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}/SHA256SUMS.asc"
curl --retry 5 -SLO "https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}/bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz"

## get Bitcoin Core signing key, verify sha256 checksum of applications and signature of SHA256SUMS.asc
gpg --import /opt/shift/config/signatures/laanwj-releases.asc
gpg --verify SHA256SUMS.asc || exit 1
grep "bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz\$" SHA256SUMS.asc | sha256sum -c - || exit 1
tar --strip-components 1 -xzf bitcoin-${BITCOIN_VERSION}-aarch64-linux-gnu.tar.gz
install -m 0755 -o root -g root -t /usr/bin bin/*

mkdir -p /etc/bitcoin/
generateConfig "bitcoin.conf.template" # --> /etc/bitcoin/bitcoin.conf
chown -R root:bitcoin /etc/bitcoin
chmod -R u+rw,g+r,g-w,o-rwx /etc/bitcoin
importFile "/etc/systemd/system/bitcoind.service"
systemctl enable bitcoind.service

redis-cli SET bitcoind:version "${BITCOIN_VERSION}"


# LIGHTNING --------------------------------------------------------------------
BIN_DEPS_TAG="v0.0.1-alpha"
LIGHTNING_VERSION_BUILD="0.7.2.1"
LIGHTNING_VERSION_BIN="0.7.0"

apt install -y libsodium-dev

## either compile c-lightning from source (default), or use prebuilt binary
if [ "${BASE_BUILD_LIGHTNINGD}" == "true" ]; then
  apt install -y  autoconf automake build-essential git libtool libgmp-dev \
                  libsqlite3-dev python python3 python3-mako net-tools \
                  zlib1g-dev asciidoc-base

  rm -rf /usr/local/src/lightning

  cd /usr/local/src/
  git clone https://github.com/ElementsProject/lightning.git
  cd lightning
  git checkout v${LIGHTNING_VERSION_BUILD}
  ./configure
  make -j 4
  make install

  redis-cli SET lightningd:version "${LIGHTNING_VERSION_BUILD}"

else
  cd /usr/local/src/
  ## temporary storage of 'lightningd' until official arm64 binaries work with stable Armbian release
  curl --retry 5 -SLO https://github.com/digitalbitbox/bitbox-base-deps/releases/download/${BIN_DEPS_TAG}/lightningd_${LIGHTNING_VERSION_BIN}-1_arm64.deb
  if ! echo "52be094f8162749acb207bf9ad08125d25288a9d03eb25690f364ba42fcff3c4  lightningd_0.7.0-1_arm64.deb" | sha256sum -c -; then
    echo "sha256sum for precompiled 'lightningd' failed" >&2
    exit 1
  fi
  dpkg -i lightningd_${LIGHTNING_VERSION_BIN}-1_arm64.deb

  ## symlink is needed, as the direct compilation (default) installs into /usr/local/bin, while this package uses '/usr/bin'
  ln -sf /usr/bin/lightningd /usr/local/bin/lightningd

  redis-cli SET lightningd:version "${LIGHTNING_VERSION_BIN}"

fi

mkdir -p /etc/lightningd/
generateConfig "lightningd.conf.template" # --> /etc/lightningd/lightningd.conf
chown -R root:bitcoin /etc/lightningd
chmod -R u+rw,g+r,g-w,o-rwx /etc/lightningd
importFile "/etc/systemd/system/lightningd.service"
systemctl enable lightningd.service


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
generateConfig "electrs.conf.template" # --> /etc/electrs/electrs.conf
chown -R root:bitcoin /etc/electrs
chmod -R u+rw,g+r,g-w,o-rwx /etc/electrs
importFile "/etc/systemd/system/electrs.service"
systemctl enable electrs.service

redis-cli SET electrs:version "${ELECTRS_VERSION}"


# TOOLS & MIDDLEWARE -------------------------------------------------------------------
BIN_DEPS_TAG="v0.0.2-alpha"

## bbbfancontrol
## see https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbfancontrol
if [ -f /opt/shift/bin/go/bbbfancontrol ]; then
  cp /opt/shift/bin/go/bbbfancontrol /usr/local/sbin/
else
  echo "INFO: bbbfancontrol not found, executable is downloaded from GitHub."
  mkdir -p /usr/local/src/bbbfancontrol && cd "$_"
  curl --retry 5 -SLO https://github.com/digitalbitbox/bitbox-base-deps/releases/download/${BIN_DEPS_TAG}/bbbfancontrol.tar.gz
  #TODO(Stadicus): add checksum once releases are stable
  # if ! echo "xxxxx  bbbfancontrol.tar.gz" | sha256sum -c -; then
  #   echo "sha256sum for precompiled 'bbbfancontrol' failed" >&2
  #   exit 1
  # fi
  tar -xzf bbbfancontrol.tar.gz
  cp ./bbbfancontrol /usr/local/sbin/
fi
importFile "/etc/systemd/system/bbbfancontrol.service"
systemctl enable bbbfancontrol.service

## bbbsupervisor
## see https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbsupervisor
if [ -f /opt/shift/bin/go/bbbsupervisor ]; then
  cp /opt/shift/bin/go/bbbsupervisor /usr/local/sbin/
  importFile "/etc/systemd/system/bbbsupervisor.service"
  systemctl enable bbbsupervisor.service
else
  #TODO(Stadicus): for ondevice build, retrieve binary from GitHub release
  echo "WARN: bbbsupervisor not found."
fi

## bbbmiddleware
## see https://github.com/digitalbitbox/bitbox-base/tree/master/middleware
if [ -f /opt/shift/bin/go/bbbmiddleware ]; then
  cp /opt/shift/bin/go/bbbmiddleware /usr/local/sbin/
  mkdir -p /etc/bbbmiddleware/
  generateConfig "bbbmiddleware.conf.template" # --> /etc/bbbmiddleware/bbbmiddleware.conf
  chmod -R u+rw,g+r,g-w,o-rwx /etc/bbbmiddleware
  importFile "/etc/systemd/system/bbbmiddleware.service"
  systemctl enable bbbmiddleware.service
else
  #TODO(Stadicus): for ondevice build, retrieve binary from GitHub release
  echo "WARN: bbbmiddleware not found."
fi


# PROMETHEUS -------------------------------------------------------------------
PROMETHEUS_VERSION="2.11.1"
PROMETHEUS_CHKSUM="33b4763032e7934870721ca3155a8ae0be6ed590af5e91bf4d2d4133a79e4548"

## Prometheus
mkdir -p /usr/local/src/prometheus && cd "$_"
curl --retry 5 -SLO https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-arm64.tar.gz
if ! echo "${PROMETHEUS_CHKSUM}  prometheus-${PROMETHEUS_VERSION}.linux-arm64.tar.gz" | sha256sum -c -; then exit 1; fi
tar --strip-components 1 -xzf prometheus-${PROMETHEUS_VERSION}.linux-arm64.tar.gz

mkdir -p /etc/prometheus /var/lib/prometheus
cp prometheus promtool /usr/local/bin/
cp -r consoles/ console_libraries/ /etc/prometheus/
chown -R prometheus /etc/prometheus /var/lib/prometheus

importFile "/etc/prometheus/prometheus.yml"
importFile "/etc/systemd/system/prometheus.service"
systemctl enable prometheus.service

## Prometheus Node Exporter
NODE_EXPORTER_VERSION="0.18.1"
NODE_EXPORTER_CHKSUM="d5a28c46e74f45b9f2158f793a6064fd9fe8fd8da6e0d1e548835ceb7beb1982"

curl --retry 5 -SLO https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/node_exporter-${NODE_EXPORTER_VERSION}.linux-arm64.tar.gz
if ! echo "${NODE_EXPORTER_CHKSUM}  node_exporter-${NODE_EXPORTER_VERSION}.linux-arm64.tar.gz" | sha256sum -c -; then exit 1; fi
tar --strip-components 1 -xzf node_exporter-${NODE_EXPORTER_VERSION}.linux-arm64.tar.gz
cp node_exporter /usr/local/bin

importFile "/etc/systemd/system/prometheus-node-exporter.service"
systemctl enable prometheus-node-exporter.service

## Prometheus Base status exporter
importFile "/etc/systemd/system/prometheus-base.service"
systemctl enable prometheus-base.service

## Prometheus Bitcoin Core exporter
importFile "/etc/systemd/system/prometheus-bitcoind.service"
systemctl enable prometheus-bitcoind.service

## Prometheus plugin for c-lightning
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

mv /etc/grafana/grafana.ini /etc/grafana/grafana.ini.default
generateConfig "grafana.ini.template"

importFile "/etc/grafana/dashboards/grafana_bitbox_base.json"
importFile "/etc/grafana/provisioning/datasources/prometheus.yaml"
importFile "/etc/grafana/provisioning/dashboards/bitbox-base.yaml"

# mkdir -p /etc/systemd/system/grafana-server.service.d/
importFile "/etc/systemd/system/grafana-server.service.d/override.conf"
systemctl enable grafana-server.service


# NGINX ------------------------------------------------------------------------
apt install -y nginx
rm -f /etc/nginx/sites-enabled/default

importFile "/etc/nginx/nginx.conf"
importFile "/etc/nginx/sites-available/grafana.conf"

if [[ "${BASE_DASHBOARD_WEB_ENABLED}" == "true" ]]; then
  /opt/shift/scripts/bbb-config.sh enable dashboard_web
fi

# mkdir -p /etc/systemd/system/nginx.service.d/
importFile "/etc/systemd/system/nginx.service.d/override.conf"
systemctl enable nginx.service

# DASHBOARD OVER HDMI ----------------------------------------------------------
mkdir -p /etc/systemd/system/getty@tty1.service.d/

if [[ "${BASE_HDMI_BUILD}" == "true" ]]; then
  apt-get install -y --no-install-recommends xserver-xorg x11-xserver-utils xinit openbox chromium
  importFile "/etc/xdg/openbox/autostart"

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
importFile "/etc/systemd/resolved.conf"

## include Wifi credentials, if specified (experimental)
if [[ -n "${BASE_WIFI_SSID}" ]]; then
  generateConfig "wlan0.conf.template" # --
  redis-cli SET network:wifi:enabled 1
fi

## mDNS services
sed -i '/PUBLISH-WORKSTATION/Ic\publish-workstation=yes' /etc/avahi/avahi-daemon.conf
generateConfig "bitboxbase.service.template" # --> /etc/avahi/services/bitboxbase.service

## firewall: restore iptables rules on startup
generateConfig "iptables.rules.template" # -->  /etc/iptables/iptables.rules
importFile "/etc/systemd/system/iptables-restore.service"


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

## Enable system services
systemctl daemon-reload
systemctl enable systemd-networkd.service
systemctl enable systemd-resolved.service
systemctl enable systemd-timesyncd.service
systemctl enable iptables-restore.service

## Set to testnet if configured
if [ "${BASE_BITCOIN_NETWORK}" == "testnet" ]; then
  /opt/shift/scripts/bbb-config.sh set bitcoin_network testnet
fi

if [ "${BASE_AUTOSETUP_SSD}" == "true" ]; then
  /opt/shift/scripts/bbb-config.sh enable autosetup_ssd
fi

## Freeze /rootfs with overlayroot (Ubuntu only)
if [ "${BASE_OVERLAYROOT}" == "true" ]; then
  if [ "${BASE_DISTRIBUTION}" == "bionic" ]; then
    echo 'overlayroot="tmpfs:swap=1,recurse=0"' > /etc/overlayroot.local.conf
  else
    echo "ERR: overlayroot is only supported in Ubuntu Bionic."
  fi
fi

## remove temporary symlink /data --> /data_source, unless building on the device
if [[ "${BASE_BUILDMODE}" != "ondevice" ]]; then
  rm /data
fi

redis-cli shutdown save

set +x
if [[ "${BASE_BUILDMODE}" == "ondevice" ]]; then
  echoArguments "Setup script finished. Please reboot device and login as user 'base'."
else
  echoArguments "Armbian build process finished. Login using SSH Keys or root password."
fi
