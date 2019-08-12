#!/bin/bash
set -eu

# BitBox Base: system configuration utility
#

SYSCONFIG_PATH="/data/sysconfig"
mkdir -p "$SYSCONFIG_PATH"

function usage() {
  echo "BitBox Base: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

possible commands:
  enable    <dashboard_hdmi|dashboard_web|wifi|autosetup_ssd|
             tor_ssh|tor_electrum|overlayroot>

  disable   any 'enable' argument

  set       <bitcoin_network|hostname|root_pw|wifi_ssid|wifi_pw>
            bitcoin_network     <mainnet|testnet>
            bitcoin_ibd         <true|false>
            bitcoin_dbcache     int (MB)
            other arguments     string

  get       any 'enable' or 'set' argument, or
            <all|tor_ssh_onion|tor_electrum_onion>

  apply     no argument, applies all configuration settings to the system 
            [not yet implemented]

  exec      bitcoin_reindex   (wipes UTXO set and validates existing blocks)
            bitcoin_resync    (re-download and validate all blocks)

"
}

if [[ ${#} -eq 0 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
  usage
  exit 0
elif [[ "${1}" == "-v" ]] || [[ "${1}" == "--version" ]]; then
  echo "bbb-config version 0.1"
  exit 0
elif [[ ${#} -eq 1 ]]; then
  usage
  exit 0
fi

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

COMMAND="${1}"
SETTING="${2^^}"

case "${COMMAND}" in
    enable|disable)
        if [[ "${COMMAND}" == "enable" ]]; then
            ENABLE=1
        else
            ENABLE=0
        fi

        case "${SETTING}" in
            DASHBOARD_HDMI)
                # enable / disable auto-login for user "hdmi", start / kill xserver
                # TODO(Stadicus): run in overlayroot-chroot for readonly rootfs
                if [[ ${ENABLE} -eq 1 ]]; then
                    mkdir -p /etc/systemd/system/getty@tty1.service.d/
                    cp /opt/shift/config/grafana/getty-override.conf /etc/systemd/system/getty@tty1.service.d/override.conf
                else
                    rm -f /etc/systemd/system/getty@tty1.service.d/override.conf
                fi
                systemctl daemon-reload
                systemctl restart getty@tty1.service
                echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                echo "Changes take effect on next restart."
                ;;

            DASHBOARD_WEB)
                # create / delete symlink to enable NGINX block
                if [[ ${ENABLE} -eq 1 ]]; then
                    ln -sf /etc/nginx/sites-available/grafana.conf /etc/nginx/sites-enabled/grafana.conf
                else
                    rm -f /etc/nginx/sites-enabled/grafana.conf
                fi
                echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                systemctl restart nginx.service
                ;;

            WIFI)
                # copy / delete wlan0 config to include directory
                if [[ ${ENABLE} -eq 1 ]]; then
                    cp /opt/shift/config/wifi/wlan0.conf /etc/network/interfaces.d/
                else
                    rm -f /etc/network/interfaces.d/wlan0.conf
                fi
                echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                systemctl restart networking.service
                ;;

            AUTOSETUP_SSD)
                echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                ;;

            TOR_SSH|TOR_ELECTRUM)
                # get service name after '_'
                SERVICE="${SETTING#*_}"

                if [[ ${ENABLE} -eq 1 ]]; then
                    # uncomment line that contains #SSH#
                    sed -i "/^#.*#${SERVICE}#/s/^#//" /etc/tor/torrc
                else
                    # comment line that contains #SERVICE# (if not commented out already)
                    sed -i "/^[^#]/ s/\(^.*#${SERVICE}#.*$\)/#\1/" /etc/tor/torrc
                fi
                echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                systemctl restart tor.service
                ;;

            OVERLAYROOT)
                if [[ ${ENABLE} -eq 1 ]]; then
                    echo 'overlayroot="tmpfs:swap=1,recurse=0"' > /etc/overlayroot.local.conf
                    echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                    echo "Overlay root filesystem will be enabled on next boot."
                else
                    overlayroot-chroot /bin/bash -c "echo 'overlayroot=disabled' > /etc/overlayroot.local.conf"
                    echo "${SETTING}=${ENABLE}" > "${SYSCONFIG_PATH}/${SETTING}"
                    echo "Overlay root filesystem will be disabled on next boot."
                fi
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
                exit 1
        esac
        cat "${SYSCONFIG_PATH}/${SETTING}"
        ;;

    set)
        if [[ ${#} -lt 3 ]]; then
            echo "Missing argument: command 'set' needs two arguments."
            exit 1
        fi

        case "${SETTING}" in
            BITCOIN_NETWORK)
                case "${3}" in
                    mainnet)
                        sed -i "/ALIAS LCLI=/Ic\alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning'" /home/base/.bashrc-custom
                        sed -i '/HIDDENSERVICEPORT 18333/Ic\HiddenServicePort 8333 127.0.0.1:8333' /etc/tor/torrc
                        sed -i '/TESTNET=/Ic\#testnet=1' /etc/bitcoin/bitcoin.conf
                        sed -i '/NETWORK=/Ic\network=bitcoin' /etc/lightningd/lightningd.conf
                        sed -i '/BITCOIN-RPCPORT=/Ic\bitcoin-rpcport=8332' /etc/lightningd/lightningd.conf
                        sed -i '/LIGHTNING-DIR=/Ic\lightning-dir=/mnt/ssd/bitcoin/.lightning' /etc/lightningd/lightningd.conf
                        sed -i '/NETWORK=/Ic\NETWORK=mainnet' /etc/electrs/electrs.conf
                        sed -i '/RPCPORT=/Ic\RPCPORT=8332' /etc/electrs/electrs.conf
                        sed -i '/BITCOIN_RPCPORT=/Ic\BITCOIN_RPCPORT=8332' /etc/bbbmiddleware/bbbmiddleware.conf || true
                        sed -i '/LIGHTNING_RPCPATH=/Ic\LIGHTNING_RPCPATH=/mnt/ssd/bitcoin/.lightning/lightning-rpc' /etc/bbbmiddleware/bbbmiddleware.conf || true
                        echo "BITCOIN_NETWORK=mainnet" > "${SYSCONFIG_PATH}/${SETTING}"
                        ;;

                    testnet)
                        sed -i "/ALIAS LCLI=/Ic\alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet'" /home/base/.bashrc-custom
                        sed -i '/HIDDENSERVICEPORT 8333/Ic\HiddenServicePort 18333 127.0.0.1:18333' /etc/tor/torrc
                        sed -i '/TESTNET=/Ic\testnet=1' /etc/bitcoin/bitcoin.conf
                        sed -i '/NETWORK=/Ic\network=testnet' /etc/lightningd/lightningd.conf
                        sed -i '/LIGHTNING-DIR=/Ic\lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet' /etc/lightningd/lightningd.conf
                        sed -i '/BITCOIN-RPCPORT=/Ic\bitcoin-rpcport=18332' /etc/lightningd/lightningd.conf
                        sed -i '/NETWORK=/Ic\NETWORK=testnet' /etc/electrs/electrs.conf
                        sed -i '/RPCPORT=/Ic\RPCPORT=18332' /etc/electrs/electrs.conf
                        sed -i '/BITCOIN_RPCPORT=/Ic\BITCOIN_RPCPORT=18332' /etc/bbbmiddleware/bbbmiddleware.conf || true
                        sed -i '/LIGHTNING_RPCPATH=/Ic\LIGHTNING_RPCPATH=/mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc' /etc/bbbmiddleware/bbbmiddleware.conf || true
                        echo "BITCOIN_NETWORK=testnet" > "${SYSCONFIG_PATH}/${SETTING}"
                        ;;

                    *)
                        echo "Invalid argument: ${SETTING} can only be set to 'mainnet' or 'testnet'."
                        exit 1
                esac
                echo "System configuration ${SETTING} will be enabled on next boot."
                ;;

            BITCOIN_IBD)
                case "${3}" in
                    true)
                        echo "Setting bitcoind configuration for 'active initial sync'."
                        bbb-config.sh set bitcoin_dbcache 2000
                        rm -f /data/triggers/bitcoind_fully_synced
                        echo "Service 'lightningd' and 'electrs' are being stopped..."
                        systemctl stop lightningd.service
                        systemctl stop electrs.service
                        ;;

                    false)
                        echo "Setting bitcoind configuration for 'fully synced'."
                        bbb-config.sh set bitcoin_dbcache 300
                        touch /data/triggers/bitcoind_fully_synced
                        echo "Service 'lightningd' and 'electrs' are being started..."
                        systemctl start lightningd.service
                        systemctl start electrs.service
                        ;;

                    *)
                        echo "Invalid argument: '${3}' must be either 'true' or 'false'."
                        exit 1
                        ;;
                esac
                ;;

            BITCOIN_DBCACHE)
                if [[ "${3}" -ge 50 ]] && [[ "${3}" -le 3000 ]]; then
                    # configure bitcoind
                    sed -i "/DBCACHE=/Ic\dbcache=${3}" /etc/bitcoin/bitcoin.conf

                    # check if service restart is necessary
                    BITCOIN_DBCACHE=0
                    source "${SYSCONFIG_PATH}/BITCOIN_DBCACHE" || true
                    if [[ "${3}" -ne "${BITCOIN_DBCACHE}" ]]; then
                        echo "Service 'bitcoind' is being restarted..."
                        systemctl restart bitcoind
                    fi
                    echo "BITCOIN_DBCACHE=${3}" > "${SYSCONFIG_PATH}/${SETTING}"
                else
                    echo "Invalid argument: '${3}' must be an integer in MB between 50 and 3000."
                    exit 1
                fi
                cat "${SYSCONFIG_PATH}/${SETTING}"
                ;;

            HOSTNAME)
                case "${3}" in
                    [^0-9A-Za-z]*|*[^\-0-9A-Z_a-z]*|*[^0-9A-Za-z]|*-_*|*_-*)
                        echo "Invalid argument: '${3}' is not a valid hostname."
                        exit 1
                        ;;
                    *)
                        echo "${3}" > /etc/hostname
                        hostname -F /etc/hostname
                        echo "${SETTING}=${3}" > "${SYSCONFIG_PATH}/${SETTING}"
                        cat "${SYSCONFIG_PATH}/${SETTING}"
                esac
                ;;

            ROOT_PW)
                # TODO(Stadicus): run in overlayroot-chroot for readonly rootfs
                echo "root:${3}" | chpasswd
                ;;

            WIFI_SSID)
                sed -i "/WPA-SSID/Ic\  wpa-ssid ${BASE_WIFI_SSID}" /opt/shift/config/wifi/wlan0.conf
                cat "${SYSCONFIG_PATH}/${SETTING}"
                ;;

            WIFI_PW)
                sed -i "/WPA-PSK/Ic\  wpa-psk ${BASE_WIFI_PW}" /opt/shift/config/wifi/wlan0.conf
                cat "${SYSCONFIG_PATH}/${SETTING}"
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."

        esac
        ;;

    get)
        case "${SETTING}" in
            BITCOIN_NETWORK|HOSTNAME|WIFI_SSID|WIFI_PW|DASHBOARD_HDMI|DASHBOARD_WEB|WIFI|AUTOSETUP_SSD|TOR_SSH|TOR_ELECTRUM)
                if [[ -f "${SYSCONFIG_PATH}/${SETTING}" ]] ; then
                    cat "${SYSCONFIG_PATH}/${SETTING}"
                else
                    echo "Missing setting, value not yet stored in configuration."
                    exit 1
                fi
                ;;

            ALL)
                cat ${SYSCONFIG_PATH}/*
                ;;

            TOR_SSH_ONION)
                echo "${SETTING}=$(cat /var/lib/tor/hidden_service_ssh/hostname)"
                ;;

            TOR_ELECTRUM_ONION)
                echo "${SETTING}=$(cat /var/lib/tor/hidden_service_electrum/hostname)"
                ;;

            ROOT_PW)
                echo "The root password is stored encrypted and cannot be provided."
                exit 1
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
        esac
        ;;

    exec)
        case "${SETTING}" in
            BITCOIN_REINDEX|BITCOIN_RESYNC)
                # stop systemd services
                systemctl stop electrs
                systemctl stop lightningd
                systemctl stop bitcoind

                if ! /bin/systemctl -q is-active bitcoind.service; then 
                    # deleting bitcoind chainstate in /mnt/ssd/bitcoin/.bitcoin/chainstate
                    rm -rf /mnt/ssd/bitcoin/.bitcoin/chainstate
                    rm -rf /mnt/ssd/electrs/db
                    rm -rf /data/triggers/bitcoind_fully_synced

                    # for RESYNC incl. download, delete `blocks` directory too
                    if [[ "${SETTING}" == "BITCOIN_RESYNC" ]]; then
                        rm -rf /mnt/ssd/bitcoin/.bitcoin/blocks

                    # otherwise assume REINDEX (only validation, no download), set option reindex-chainstate
                    else
                        echo "reindex-chainstate=1" >> /etc/bitcoin/bitcoin.conf

                    fi

                    # restart bitcoind and remove option
                    systemctl start bitcoind
                    sleep 10
                    sed -i '/reindex/Id' /etc/bitcoin/bitcoin.conf

                else
                    echo "bitcoind is still running, cannot delete chainstate"
                    exit 1
                fi

                echo "Command ${SETTING} successfully executed."
                ;;

            *)
                echo "Invalid argument: exec command ${SETTING} unknown."
        esac
        ;;

    *)
        echo "Invalid argument: command ${COMMAND} unknown."
        exit 1
esac