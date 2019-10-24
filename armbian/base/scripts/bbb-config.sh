#!/bin/bash
# shellcheck disable=SC1091
set -eu

# BitBox Base: system configuration utility
#

# print usage information for script
usage() {
  echo "BitBox Base: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

assumes Redis database running to be used with 'redis-cli'

possible commands:
  enable    <bitcoin_incoming|bitcoin_ibd|bitcoin_ibd_clearnet|dashboard_hdmi|
             dashboard_web|wifi|autosetup_ssd|tor|tor_bbbmiddleware|tor_ssh|
             tor_electrum|overlayroot|pwlogin|rootlogin|unsigned_updates>

  disable   any 'enable' argument

  set       <hostname|root_pw|wifi_ssid|wifi_pw>
            bitcoin_network         <mainnet|testnet>
            bitcoin_dbcache         int (MB)
            other arguments         string

"
}

# MockMode checks all arguments but does not execute anything
#
# usage: call this script with the ENV variable MOCKMODE set to 1, e.g.
#        $ MOCKMODE=1 ./bbb-config.sh
#
MOCKMODE=${MOCKMODE:-0}
checkMockMode() {
    if [[ $MOCKMODE -eq 1 ]]; then
        echo "MOCK MODE enabled"
        echo "OK: ${COMMAND} -- ${SETTING}"
        exit 0
    fi
}

# error handling
errorExit() {
    echo "$@" 1>&2
    exit 1
}

# don't load includes for MockMode
if [[ $MOCKMODE -ne 1 ]]; then

    if [[ ! -d /opt/shift/scripts/include/ ]]; then
        echo "ERR: includes directory /opt/shift/scripts/include/ not found, must run on BitBox Base system. Run in MockMode for testing."
        errorExit SCRIPT_INCLUDES_NOT_FOUND
    fi

    # include function exec_overlayroot(), to execute a command, either within overlayroot-chroot or directly
    source /opt/shift/scripts/include/exec_overlayroot.sh.inc

    # include functions redis_set() and redis_get()
    source /opt/shift/scripts/include/redis.sh.inc

    # include function generateConfig() to generate config files from templates
    source /opt/shift/scripts/include/generateConfig.sh.inc
fi

# include updateTorOnions() function
source /opt/shift/scripts/include/updateTorOnions.sh.inc

# ------------------------------------------------------------------------------


# check script arguments
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

if [[ $MOCKMODE -ne 1 ]] && [[ ${UID} -ne 0 ]]; then
    echo "${0}: needs to be run as superuser." >&2
    errorExit SCRIPT_NOT_RUN_AS_SUPERUSER
fi

COMMAND="${1}"
SETTING="${2^^}"

# parse COMMAND: enable, disable, get, set
case "${COMMAND}" in
    enable|disable)
        if [[ "${COMMAND}" == "enable" ]]; then
            ENABLE=1
        else
            ENABLE=0
        fi

        case "${SETTING}" in
            BITCOIN_INCOMING)
                checkMockMode

                redis_set "bitcoind:listen" "${ENABLE}"
                generateConfig "bitcoin.conf.template"
                systemctl restart bitcoind.service
                ;;

            BITCOIN_IBD)
                checkMockMode

                if [[ ${ENABLE} -eq 1 ]]; then
                    echo "Service 'lightningd' and 'electrs' are being stopped..."
                    systemctl stop lightningd.service
                    systemctl stop electrs.service
                    echo "Setting bitcoind configuration for 'active initial sync'."
                    bbb-config.sh set bitcoin_dbcache 2000
                    redis_set "bitcoind:ibd" "${ENABLE}"

                else
                    echo "Setting bitcoind configuration for 'fully synced'."
                    bbb-config.sh set bitcoin_dbcache 300
                    redis_set "bitcoind:ibd" "${ENABLE}"
                    echo "Service 'lightningd' and 'electrs' are being started..."
                    systemctl start lightningd.service
                    systemctl start electrs.service
                fi
                ;;

            BITCOIN_IBD_CLEARNET)
                checkMockMode

                # don't set option if Tor is disabled globally
                if [[ ${ENABLE} -eq 1 ]] && [ "$(redis_get 'tor:base:enabled')" -eq 0 ]; then
                    echo "ERR: Tor service is already disabled for the whole system, cannot enable BITCOIN_IBD_CLEARNET"
                    errorExit ENABLE_CLEARNETIBD_TOR_ALREADY_DISABLED
                fi

                # configure bitcoind to run over IPv4 while in IBD mode
                redis_set "bitcoind:ibd-clearnet" "${ENABLE}"
                generateConfig "iptables.rules.template"
                generateConfig "bitcoin.conf.template"
                systemctl start iptables-restore
                systemctl restart bitcoind
                echo "OK: bitcoind:ibd-clearnet set to ${ENABLE}"
                ;;

            DASHBOARD_HDMI)
                checkMockMode

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
                redis_set "base:dashboard:hdmi:enabled" "${ENABLE}"
                echo "Changes take effect on next restart."
                ;;

            DASHBOARD_WEB)
                checkMockMode

                # create / delete symlink to enable NGINX block
                # TODO(Stadicus): run in overlayroot-chroot for readonly rootfs
                if [[ ${ENABLE} -eq 1 ]]; then
                    ln -sf /etc/nginx/sites-available/grafana.conf /etc/nginx/sites-enabled/grafana.conf
                    systemctl enable grafana-server.service
                    systemctl start grafana-server.service
                else
                    rm -f /etc/nginx/sites-enabled/grafana.conf
                    systemctl disable grafana-server.service
                    systemctl stop grafana-server.service
                fi
                redis_set "base:dashboard:web:enabled" "${ENABLE}"
                systemctl restart nginx.service
                ;;

            WIFI)
                checkMockMode

                # copy / delete wlan0 config to include directory
                # TODO(Stadicus): run in overlayroot-chroot for readonly rootfs

                if [[ ${ENABLE} -eq 1 ]]; then
                    generateConfig "wlan0.conf.template"
                else
                    rm -f /etc/network/interfaces.d/wlan0.conf
                fi
                redis_set "base:wifi:enabled" "${ENABLE}"
                systemctl restart networking.service
                ;;

            AUTOSETUP_SSD)
                checkMockMode

                if [[ ${ENABLE} -eq 1 ]]; then
                    touch /opt/shift/config/.autosetup-ssd
                else
                    exec_overlayroot all-layers "rm /opt/shift/config/.autosetup-ssd"
                fi
                redis_set "base:autosetupssd:enabled" "${ENABLE}"
                ;;

            TOR)
                checkMockMode

                # enable/disable Tor systemwide
                if [[ ${ENABLE} -eq 1 ]]; then
                    exec_overlayroot all-layers "systemctl enable tor.service"
                    redis_set "tor:base:enabled" "${ENABLE}"
                    generateConfig "iptables.rules.template"
                    generateConfig "bitcoin.conf.template"
                    generateConfig "lightningd.conf.template"
                    systemctl start tor.service
                else
                    exec_overlayroot all-layers "systemctl disable tor.service"
                    redis_set "tor:base:enabled" "${ENABLE}"
                    generateConfig "iptables.rules.template"
                    generateConfig "bitcoin.conf.template"
                    generateConfig "lightningd.conf.template"
                    systemctl stop tor.service
                fi
                echo "Restarting services..."
                systemctl start iptables-restore
                systemctl restart networking.service
                systemctl restart bitcoind.service
                systemctl restart lightningd.service || true        # allowed to fail if bitcoind is in IBD mode
                redis_set "tor:base:enabled" "${ENABLE}"
                updateTorOnions
                ;;

            TOR_SSH|TOR_ELECTRS|TOR_BBBMIDDLEWARE)
                if [[ ${SETTING} == "TOR_SSH" ]]; then
                    checkMockMode
                    redis_set "tor:ssh:enabled" "${ENABLE}"
                elif [[ ${SETTING} == "TOR_ELECTRS" ]]; then
                    redis_set "tor:electrs:enabled" "${ENABLE}"
                elif [[ ${SETTING} == "TOR_BBBMIDDLEWARE" ]]; then
                    checkMockMode
                    redis_set "tor:bbbmiddleware:enabled" "${ENABLE}"
                else
                    echo "ERR: invalid argument, setting ${SETTING} not allowed"
                    errorExit CONFIG_SCRIPT_INVALID_ARG
                fi

                generateConfig "torrc.template"
                systemctl restart tor.service

                updateTorOnions
                ;;

            OVERLAYROOT)
                checkMockMode

                # set explicitly without exec_overlayroot() to make sure it is set under all conditions
                if [[ ${ENABLE} -eq 1 ]]; then
                    echo 'overlayroot="tmpfs:swap=1,recurse=0"' > /etc/overlayroot.local.conf
                    echo "Overlay root filesystem will be enabled on next boot."
                else
                    overlayroot-chroot /bin/bash -c "echo 'overlayroot=disabled' > /etc/overlayroot.local.conf"
                    echo "Overlay root filesystem will be disabled on next boot."
                fi
                redis_set "base:overlayroot:enabled" "${ENABLE}"
                ;;

            ROOTLOGIN)
                # internal use only, eg. to allow scp/ftp during development
                # option is not meant to be available in user interface
                checkMockMode

                if [[ ${ENABLE} -eq 1 ]]; then
                    redis_set "base:sshd:rootlogin" "yes"
                else
                    redis_set "base:sshd:rootlogin" "no"
                fi
                generateConfig "sshd_config.template"
                systemctl restart sshd.service
                ;;

            PWLOGIN)
                checkMockMode

                if [[ ${ENABLE} -eq 1 ]]; then
                    redis_set "base:sshd:passwordlogin" "yes"
                else
                    redis_set "base:sshd:passwordlogin" "no"
                fi
                generateConfig "sshd_config.template"
                systemctl restart sshd.service
                ;;

            UNSIGNED_UPDATES)
                checkMockMode

                redis_set "base:update:allow-unsigned" "${ENABLE}"
                generateConfig "mender.conf.template"
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
                errorExit CONFIG_SCRIPT_INVALID_ARG
        esac
        ;;

    set)
        if [[ ${#} -lt 3 ]]; then
            echo "Missing argument: command 'set' needs two arguments."
            errorExit SET_NEEDS_TWO_ARGUMENTS
        fi

        case "${SETTING}" in
            BITCOIN_NETWORK)
                case "${3}" in
                    mainnet)
                        checkMockMode

                        redis_set "bitcoind:network" "mainnet"
                        redis_set "bitcoind:testnet" "0"
                        redis_set "bitcoind:mainnet" "1"
                        redis_set "lightningd:lightning-dir" "/mnt/ssd/bitcoin/.lightning"
                        ;;

                    testnet)
                        checkMockMode

                        redis_set "bitcoind:network" "testnet"
                        redis_set "bitcoind:testnet" "1"
                        redis_set "bitcoind:mainnet" "0"
                        redis_set "lightningd:lightning-dir" "/mnt/ssd/bitcoin/.lightning-testnet"
                        ;;

                    *)
                        echo "Invalid argument: ${SETTING} can only be set to 'mainnet' or 'testnet'."
                        errorExit SET_BITCOINETWORK_INVALID_VALUE
                esac

                generateConfig "bashrc-custom.template"
                generateConfig "torrc.template"
                generateConfig "bitcoin.conf.template"
                generateConfig "lightningd.conf.template"
                generateConfig "electrs.conf.template"
                generateConfig "bbbmiddleware.conf.template"
                echo "System configuration ${SETTING} will be enabled on next boot."
                ;;

            # DEPRECATED -- TODO(Stadicus) remove
            BITCOIN_IBD)
                case "${3}" in
                    true)
                        bbb-config.sh enable bitcoin_ibd
                        ;;

                    false)
                        bbb-config.sh disable bitcoin_ibd
                        ;;

                    *)
                        echo "Invalid argument: '${3}' must be either 'true' or 'false'."
                        errorExit SET_BITCOINIBD_INVALID_VALUE
                        ;;
                esac
                echo "DEPRECATED: please use bbb-config.sh <enable|disable> bitcoin_ibd"
                ;;

            # DEPRECATED -- TODO(Stadicus) remove
            BITCOIN_IBD_CLEARNET)
                case "${3}" in
                    true)
                        bbb-config.sh enable bitcoin_ibd_clearnet
                        ;;

                    false)
                        bbb-config.sh disable bitcoin_ibd_clearnet
                        ;;
                    *)
                        echo "ERR: argument needs to be either 'true' or 'false'"
                        errorExit SET_BITCOINIBD_CLEARNET_INVALID_VALUE
                esac
                echo "DEPRECATED: please use bbb-config.sh <enable|disable> bitcoin_ibd_clearnet"
                ;;

            BITCOIN_DBCACHE)
                if [[ "${3}" -ge 50 ]] && [[ "${3}" -le 3000 ]]; then
                    checkMockMode

                    DBCACHE_BEFORE=$(redis_get "bitcoind:dbcache")
                    redis_set "bitcoind:dbcache" "${3}"

                    generateConfig "bitcoin.conf.template"

                    # check if service restart is necessary
                    if [[ "${DBCACHE_BEFORE}" == "${3}" ]]; then
                        echo "DBCACHE unchanged (${DBCACHE_BEFORE} MB to ${3} MB), no restart of bitcoind required"
                    else
                        echo "DBCACHE changed (${DBCACHE_BEFORE} MB to ${3} MB), restarting bitcoind"
                        systemctl restart bitcoind.service
                    fi

                else
                    echo "Invalid argument: '${3}' must be an integer in MB between 50 and 3000."
                    errorExit SET_BITCOINDBCACHE_INVALID_VALUE
                fi
                ;;

            HOSTNAME)
                # check that hostname is valid
                regex='^[a-z][a-z0-9-]{0,22}[a-z0-9]$'
                if [[ "${3}" =~ ${regex} ]]; then
                    checkMockMode

                    exec_overlayroot all-layers "echo '${3}' > /etc/hostname"
                    exec_overlayroot all-layers "echo '127.0.0.1   localhost ${3}' > /etc/hosts"
                    hostname -F /etc/hostname
                    redis_set "base:hostname" "${3}"
                    systemctl restart networking.service        || true
                    systemctl restart avahi-daemon.service      || true
                else
                    echo "Invalid argument: ${3} is not a valid hostname."
                    errorExit SET_HOSTNAME_INVALID_VALUE
                fi
                ;;

            ROOT_PW)
                checkMockMode

                exec_overlayroot all-layers "echo 'root:${3}' | chpasswd"
                exec_overlayroot all-layers "echo 'base:${3}' | chpasswd"
                ;;

            WIFI_SSID)
                checkMockMode

                redis_set "base:wifi:ssid" "${3}"
                ;;

            WIFI_PW)
                checkMockMode

                redis_set "base:wifi:password" "${3}"
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
                errorExit CONFIG_SCRIPT_INVALID_ARG
        esac
        ;;

        *)
            echo "Invalid argument: command ${COMMAND} unknown."
            errorExit CONFIG_SCRIPT_INVALID_ARG

esac
