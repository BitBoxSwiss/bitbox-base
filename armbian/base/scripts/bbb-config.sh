#!/bin/bash
set -eu

# BitBox Base: system configuration utility
#

SYSCONFIG_PATH="/opt/shift/sysconfig"
mkdir -p "$SYSCONFIG_PATH"

function usage() {
  echo "BitBox Base: system configuration utility
usage: bbb-config [--version] [--help]
                  <command> [<args>]

possible commands:
  enable    <dashboard_hdmi|dashboard_web|wifi|autosetup_ssd|
             tor_ssh|tor_electrum>

  disable   any 'enable' argument

  set       <bitcoin_network|hostname|root_pw|wifi_ssid|wifi_pw>
            bitcoin_network     <mainnet|testnet>
            other arguments     string

  get       any 'enable' or 'set' argument, or
            <all|tor_ssh_onion|tor_electrum_onion>

"
}

if [[ ${#} -eq 0 ]] || [[ "${1}" == "-h" ]] || [[ "${1}" == "--help" ]]; then
  usage
  exit 0
elif [[ "${1}" == "-v" ]] || [[ "${1}" == "--version" ]]; then
  echo "bbb-config version 0.1"
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
                if [[ ${ENABLE} -eq 1 ]]; then
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

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
                exit 1
        esac
        cat "${SYSCONFIG_PATH}/${SETTING}"
        ;;

    set)
        if [[ -z ${3} ]]; then
            echo "Missing argument: command 'set' needs two arguments."
            exit 1
        fi

        case "${SETTING}" in
            BITCOIN_NETWORK)
                case "${3}" in
                    mainnet)
                        sed -i '/CONFIGURED FOR/Ic\echo "Configured for Bitcoin MAINNET"; echo' /etc/update-motd.d/20-shift
                        sed -i "/ALIAS BLOG=/Ic\alias blog='tail -f /mnt/ssd/bitcoin/.bitcoin/debug.log'" /root/.bashrc-custom
                        sed -i "/ALIAS LCLI=/Ic\alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning'" /root/.bashrc-custom
                        sed -i '/HIDDENSERVICEPORT 18333/Ic\HiddenServicePort 8333 127.0.0.1:8333' /etc/tor/torrc
                        sed -i '/TESTNET=/Ic\#testnet=1' /etc/bitcoin/bitcoin.conf
                        sed -i '/NETWORK=/Ic\network=bitcoin' /etc/lightningd/lightningd.conf
                        sed -i '/BITCOIN-RPCPORT=/Ic\bitcoin-rpcport=8332' /etc/lightningd/lightningd.conf
                        sed -i '/LIGHTNING-DIR=/Ic\lightning-dir=/mnt/ssd/bitcoin/.lightning' /etc/lightningd/lightningd.conf
                        sed -i '/NETWORK=/Ic\NETWORK=mainnet' /etc/electrs/electrs.conf
                        sed -i '/RPCPORT=/Ic\RPCPORT=8332' /etc/electrs/electrs.conf
                        sed -i '/BITCOIN_RPCPORT=/Ic\BITCOIN_RPCPORT=8332' /etc/base-middleware/base-middleware.conf
                        sed -i '/LIGHTNING_RPCPATH=/Ic\LIGHTNING_RPCPATH=/mnt/ssd/bitcoin/.lightning/lightning-rpc' /etc/base-middleware/base-middleware.conf
                        sed -i '/<PORT>18333/Ic\<port>8333</port>' /etc/avahi/services/bitcoind.service
                        echo "BITCOIN_NETWORK=mainnet" > "${SYSCONFIG_PATH}/${SETTING}"
                        ;;

                    testnet)
                        sed -i '/CONFIGURED FOR/Ic\echo "Configured for Bitcoin TESTNET"; echo' /etc/update-motd.d/20-shift
                        sed -i "/ALIAS BLOG=/Ic\alias blog='tail -f /mnt/ssd/bitcoin/.bitcoin/testnet3/debug.log'" /root/.bashrc-custom
                        sed -i "/ALIAS LCLI=/Ic\alias lcli='lightning-cli --lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet'" /root/.bashrc-custom
                        sed -i '/HIDDENSERVICEPORT 8333/Ic\HiddenServicePort 18333 127.0.0.1:18333' /etc/tor/torrc
                        sed -i '/TESTNET=/Ic\testnet=1' /etc/bitcoin/bitcoin.conf
                        sed -i '/NETWORK=/Ic\network=testnet' /etc/lightningd/lightningd.conf
                        sed -i '/LIGHTNING-DIR=/Ic\lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet' /etc/lightningd/lightningd.conf
                        sed -i '/BITCOIN-RPCPORT=/Ic\bitcoin-rpcport=18332' /etc/lightningd/lightningd.conf
                        sed -i '/NETWORK=/Ic\NETWORK=testnet' /etc/electrs/electrs.conf
                        sed -i '/RPCPORT=/Ic\RPCPORT=18332' /etc/electrs/electrs.conf
                        sed -i '/BITCOIN_RPCPORT=/Ic\BITCOIN_RPCPORT=18332' /etc/base-middleware/base-middleware.conf
                        sed -i '/LIGHTNING_RPCPATH=/Ic\LIGHTNING_RPCPATH=/mnt/ssd/bitcoin/.lightning-testnet/lightning-rpc' /etc/base-middleware/base-middleware.conf
                        sed -i '/<PORT>8333/Ic\<port>18333</port>' /etc/avahi/services/bitcoind.service
                        echo "BITCOIN_NETWORK=testnet" > "${SYSCONFIG_PATH}/${SETTING}"
                        ;;

                    *)
                        echo "Invalid argument: ${SETTING} can only be set to 'mainnet' or 'testnet'."
                        exit 1
                esac
                echo "System configuration ${SETTING} will be enabled on next boot."
                ;;

            HOSTNAME)
                case "${3}" in
                    [^0-9A-Za-z]*|*[^\-0-9A-Z_a-z]*|*[^0-9A-Za-z]|*-_*|*_-*)
                        echo "Invalid argument: '${3}' is not a valid hostname."
                        exit 1
                        ;;
                    *)
                        echo "${3}" > /etc/hostname
                        echo "${SETTING}=${3}" > "${SYSCONFIG_PATH}/${SETTING}"
                        cat "${SYSCONFIG_PATH}/${SETTING}"
                esac
                ;;

            ROOT_PW)
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

    *)
        echo "Invalid argument: command ${COMMAND} unknown."
        exit 1
esac