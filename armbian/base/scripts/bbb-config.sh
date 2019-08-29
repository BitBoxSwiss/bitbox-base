#!/bin/bash
set -eu

# BitBox Base: system configuration utility
#

# print usage information for script
usage() {
  echo "BitBox Base: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

assumes Redis database running to be used with 'redis-cli', respecting
environment variables REDIS_HOST, REDIS_PORT and REDIS_DB

possible commands:
  enable    <dashboard_hdmi|dashboard_web|wifi|autosetup_ssd|
             tor|tor_ssh|tor_electrum|overlayroot|root_pwlogin>

  disable   any 'enable' argument

  set       <bitcoin_network|hostname|root_pw|wifi_ssid|wifi_pw>
            bitcoin_network     <mainnet|testnet>
            bitcoin_ibd         <true|false>
            bitcoin_dbcache     int (MB)
            other arguments     string

  get       <tor_ssh_onion|tor_electrum_onion>

  apply     no argument, applies all configuration settings to the system 
            [not yet implemented]

"
}

# Redis configuration
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_DB="${REDIS_DB:-0}"

# function to execute command, either within overlayroot-chroot or directly
exec_overlayroot() {
    echo "${1}"
    if [[ "${1}" != "base-only" ]] && [[ "${1}" != "all-layers" ]]; then
        echo "exec_overlayroot(): first argument '${1}', but must be either"
        echo "                    'base-only':  execute base layer (in r/o partition when overlayroot active, or directy when no overlayroot active"
        echo "                    'all-layers': execute both in overlayroot and directly"
        exit 1
    fi

    if [ "${OVERLAYROOT_ENABLED}" -eq 1 ]; then
        echo "executing in overlayroot-chroot: ${2}"
        overlayroot-chroot /bin/bash -c "${2}"
    fi

    if [ "${OVERLAYROOT_ENABLED}" -ne 1 ] || [[ "${1}" == "all-layers" ]]; then
        echo "executing directly: ${2}"
        /bin/bash -c "${2}"
    fi
}

redis_set() {
    # usage: redis_set "key" "value"
    ok=$(redis-cli -h localhost -p 6379 -n 0 SET "${1}" "${2}")
    if [[ "${ok}"  != "OK" ]]; then
        echo "ERR: could not SET key ${1}"
        # exit 1
    fi
}

redis_get() {
    # usage: str=$(redis_get "key")
    ok=$(redis-cli -h localhost -p 6379 -n 0 GET "${1}")
    echo "${ok}"
}

generateConfig() {
  # generates a config file using custom bbbconfig
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
  exit 1
fi

COMMAND="${1}"
SETTING="${2^^}"

# check if overlayroot is enabled
OVERLAYROOT_ENABLED=0
if grep -q "tmpfs" /etc/overlayroot.local.conf; then
    OVERLAYROOT_ENABLED=1
fi

# parse COMMAND: enable, disable, get, set
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
                redis_set "base:dashboard:hdmi:enabled" "${ENABLE}"
                echo "Changes take effect on next restart."
                ;;

            DASHBOARD_WEB)
                # create / delete symlink to enable NGINX block
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
                if [[ ${ENABLE} -eq 1 ]]; then
                    generateConfig "wlan0.conf.template"
                else
                    rm -f /etc/network/interfaces.d/wlan0.conf
                fi
                redis_set "base:wifi:enabled" "${ENABLE}"
                systemctl restart networking.service
                ;;

            AUTOSETUP_SSD)
                redis_set "base:autosetupssd:enabled" "${ENABLE}"
                ;;

            TOR)
                # enable/disable Tor systemwide
                if [[ ${ENABLE} -eq 1 ]]; then
                    exec_overlayroot all-layers "systemctl enable tor.service"
                    redis_set "tor:base:enabled" "${ENABLE}"
                    redis_set "bitcoind:onlynet:enabled" 0
                    generateConfig "bitcoin.conf.template"
                    systemctl start tor.service
                else
                    exec_overlayroot all-layers "systemctl disable tor.service"
                    redis_set "tor:base:enabled" "${ENABLE}"
                    redis_set "bitcoind:onlynet:enabled" 1
                    generateConfig "bitcoin.conf.template"
                    systemctl stop tor.service
                fi
                systemctl restart networking.service
                systemctl restart bitcoind.service
                ;;

            TOR_SSH|TOR_ELECTRUM)
                if [[ ${SETTING} == "TOR_SSH" ]]; then
                    redis_set "base:tor:ssh:enabled" "${ENABLE}"
                elif [[ ${SETTING} == "TOR_ELECTRUM" ]]; then
                    redis_set "base:tor:electrs:enabled" "${ENABLE}"
                else
                    echo "ERR: invalid argument, setting ${SETTING} not allowed"
                    exit 1
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
                exit 1
        esac
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
                        exit 1
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
                        exit 1
                        ;;
                esac
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
                    exit 1
                fi
                ;;

            HOSTNAME)
                case "${3}" in
                    [^0-9A-Za-z]*|*[^-0-9A-Z_a-z]*|*[^0-9A-Za-z]|*-_*|*_-*)
                        echo "Invalid argument: '${3}' is not a valid hostname."
                        exit 1
                        ;;
                    *)
                        exec_overlayroot all-layers "echo '${3}' > /etc/hostname"
                        exec_overlayroot all-layers "echo '127.0.0.1   localhost ${3}' > /etc/hosts"
                        hostname -F /etc/hostname
                        redis_set "base:hostname" "${3}"
                esac
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
        esac
        ;;

    *)
        echo "Invalid argument: command ${COMMAND} unknown."
        exit 1//
esac