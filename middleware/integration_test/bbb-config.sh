#!/bin/bash
set -eu

# MOCK DEV SCRIPT
# always return without doing anything

# BitBox Base: system configuration utility
#

# print usage information for script
usage() {
  echo "BitBox Base: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

assumes Redis database running to be used with 'redis-cli'

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
            DASHBOARD_HDMI|DASHBOARD_WEB|WIFI|AUTOSETUP_SSD|TOR|TOR_SSH|TOR_ELECTRUM|OVERLAYROOT|ROOT_PWLOGIN)
                echo "OK: ${COMMAND} -- ${SETTING}"
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
                    mainnet|testnet)
                        echo "OK: ${COMMAND} -- ${SETTING} ${3}"
                        ;;

                    *)
                        echo "Invalid argument: ${SETTING} can only be set to 'mainnet' or 'testnet'."
                        exit 1
                esac
                ;;

            BITCOIN_IBD|BITCOIN_IBD_CLEARNET)
                case "${3}" in
                    true|false)
                        echo "OK: ${COMMAND} -- ${SETTING} ${3}"
                        ;;
                    *)
                        echo "Invalid argument: '${3}' must be either 'true' or 'false'."
                        exit 1
                esac
                ;;

            BITCOIN_DBCACHE)
                if [[ "${3}" -ge 50 ]] && [[ "${3}" -le 3000 ]]; then
                    echo "OK: ${COMMAND} -- ${SETTING} ${3}"
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
                        echo "OK: ${COMMAND} -- ${SETTING} ${3}"
                esac
                ;;

            ROOT_PW|WIFI_SSID|WIFI_PW)
                echo "OK: ${COMMAND} -- ${SETTING} ${3}"
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."

        esac
        ;;

    get)
        case "${SETTING}" in
            TOR_SSH_ONION)
                echo "${SETTING}=torssh-onioneiynxxxxxxxxxxxaicxqgb3xxxxxxxxxabxegcjznhyd.onion"
                ;;

            TOR_ELECTRUM_ONION)
                echo "${SETTING}=torelectrumneiynxxxxxxxxxxxaicxqgb3xxxxxxxxxabxegcjznhyd.onion"
                ;;

            *)
                echo "Invalid argument: setting ${SETTING} unknown."
        esac
        ;;

    *)
        echo "Invalid argument: command ${COMMAND} unknown."
        exit 1
esac
