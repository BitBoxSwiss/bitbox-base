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
  enable    <bitcoin_incoming|dashboard_hdmi|dashboard_web|wifi|autosetup_ssd|
             tor|tor_bbbmiddleware|tor_ssh|tor_electrum|overlayroot|root_pwlogin>

  disable   any 'enable' argument

  set       <bitcoin_network|hostname|root_pw|wifi_ssid|wifi_pw>
            bitcoin_network         <mainnet|testnet>
            bitcoin_ibd             <true|false>
            bitcoin_ibd_clearnet    <true|false>
            bitcoin_dbcache         int (MB)
            other arguments         string

  get       <tor_ssh_onion|tor_electrum_onion>

  apply     no argument, applies all configuration settings to the system 
            [not yet implemented]

"
}

# include function exec_overlayroot(), to execute a command, either within overlayroot-chroot or directly
source /opt/shift/scripts/include/exec_overlayroot.sh.inc

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# include function generateConfig() to generate config files from templates
source /opt/shift/scripts/include/generateConfig.sh.inc

# error handling function
errorExit() {
    echo "$@" 1>&2
    exit 1
}

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

if [[ ${UID} -ne 0 ]]; then
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
                redis_set "bitcoind:listen" "${ENABLE}"
                generateConfig "bitcoin.conf.template"
                systemctl restart bitcoind.service
                ;;
                
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
                redis_set "base:dashboard:hdmi:enabled" "${ENABLE}"
                echo "Changes take effect on next restart."
                ;;

            DASHBOARD_WEB)
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
                if [[ ${ENABLE} -eq 1 ]]; then
                    touch /opt/shift/config/.autosetup-ssd
                else
                    exec_overlayroot all-layers "rm /opt/shift/config/.autosetup-ssd"
                fi
                redis_set "base:autosetupssd:enabled" "${ENABLE}"
                ;;

            TOR)
                # enable/disable Tor systemwide
                if [[ ${ENABLE} -eq 1 ]]; then
                    exec_overlayroot all-layers "systemctl enable tor.service"
                    redis_set "tor:base:enabled" "${ENABLE}"
                    generateConfig "bitcoin.conf.template"
                    generateConfig "lightningd.conf.template"
                    systemctl start tor.service
                else
                    exec_overlayroot all-layers "systemctl disable tor.service"
                    redis_set "tor:base:enabled" "${ENABLE}"
                    generateConfig "bitcoin.conf.template"
                    generateConfig "lightningd.conf.template"
                    systemctl stop tor.service
                fi
                echo "Restarting services..."
                systemctl restart networking.service
                systemctl restart bitcoind.service

                redis_set "tor:base:enabled" "${ENABLE}"
                ;;

            TOR_SSH|TOR_ELECTRUM|TOR_BBBMIDDLEWARE)
                if [[ ${SETTING} == "TOR_SSH" ]]; then
                    redis_set "tor:ssh:enabled" "${ENABLE}"
                elif [[ ${SETTING} == "TOR_ELECTRUM" ]]; then
                    redis_set "tor:electrs:enabled" "${ENABLE}"
                elif [[ ${SETTING} == "TOR_BBBMIDDLEWARE" ]]; then
                    redis_set "tor:bbbmiddleware:enabled" "${ENABLE}"
                else
                    echo "ERR: invalid argument, setting ${SETTING} not allowed"
                    errorExit CONFIG_SCRIPT_INVALID_ARG
                fi

                generateConfig "torrc.template"
                systemctl restart tor.service
                ;;

            OVERLAYROOT)
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

            ROOT_PWLOGIN)
                # unlock/lock root user for password login
                if [[ ${ENABLE} -eq 1 ]]; then
                    exec_overlayroot all-layers "passwd -u root"
                else
                    exec_overlayroot all-layers "passwd -l root"
                fi
                redis_set "base:rootpasslogin:enabled" "${ENABLE}"
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
                        redis_set "bitcoind:network" "mainnet"
                        redis_set "bitcoind:testnet" "0"
                        redis_set "bitcoind:mainnet" "1"
                        redis_set "lightningd:lightning-dir" "/mnt/ssd/bitcoin/.lightning"
                        ;;

                    testnet)
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

            BITCOIN_IBD)
                case "${3}" in
                    true)
                        echo "Service 'lightningd' and 'electrs' are being stopped..."
                        systemctl stop lightningd.service
                        systemctl stop electrs.service
                        echo "Setting bitcoind configuration for 'active initial sync'."
                        bbb-config.sh set bitcoin_dbcache 2000
                        redis_set "bitcoind:ibd" "1"
                        ;;

                    false)
                        echo "Setting bitcoind configuration for 'fully synced'."
                        bbb-config.sh set bitcoin_dbcache 300
                        redis_set "bitcoind:ibd" "0"
                        echo "Service 'lightningd' and 'electrs' are being started..."
                        systemctl start lightningd.service
                        systemctl start electrs.service
                        ;;

                    *)
                        echo "Invalid argument: '${3}' must be either 'true' or 'false'."
                        errorExit SET_BITCOINIBD_INVALID_VALUE
                        ;;
                esac
                ;;

            BITCOIN_IBD_CLEARNET)
                case "${3}" in
                    true)
                        # don't set option if Tor is disabled globally
                        if [ "$(redis_get 'tor:base:enabled')" -eq 0 ]; then
                            echo "ERR: Tor service is already disabled for the whole system, cannot enable BITCOIN_IBD_CLEARNET"
                            errorExit ENABLE_CLEARNETIBD_TOR_ALREADY_DISABLED
                        fi
                        SET=1
                        ;;

                    false)
                        SET=0
                        ;;
                    *)
                        echo "ERR: argument needs to be either 'true' or 'false'"
                        errorExit SET_BITCOINIBD_CLEARNET_INVALID_VALUE
                esac

                # configure bitcoind to run over IPv4 while in IBD mode
                redis_set "bitcoind:ibd-clearnet" "${SET}"
                generateConfig "bitcoin.conf.template"
                systemctl restart bitcoind
                echo "OK: bitcoind:ibd-clearnet set to ${SET}"
                ;;

            BITCOIN_DBCACHE)
                if [[ "${3}" -ge 50 ]] && [[ "${3}" -le 3000 ]]; then

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
                exec_overlayroot all-layers "echo 'root:${3}' | chpasswd"
                exec_overlayroot all-layers "echo 'base:${3}' | chpasswd"
                ;;

            WIFI_SSID)
                redis_set "base:wifi:ssid" "${3}"
                ;;

            WIFI_PW)
                redis_set "base:wifi:password" "${3}"
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
                errorExit CONFIG_SCRIPT_INVALID_ARG
        esac
        ;;

    get)
        case "${SETTING}" in
            TOR_SSH_ONION)
                echo "${SETTING}=$(cat /var/lib/tor/hidden_service_ssh/hostname)"
                ;;

            TOR_ELECTRUM_ONION)
                echo "${SETTING}=$(cat /var/lib/tor/hidden_service_electrum/hostname)"
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