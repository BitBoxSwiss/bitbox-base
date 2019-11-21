#!/bin/bash
# shellcheck disable=SC1091

# This script is called by the update-checks.service on boot
# to check if update is in progress and perform configuration and
# validation actions.
#
# The Redis config mgmt must available for this script.
#
set -eu

# --- generic functions --------------------------------------------------------

# include function exec_overlayroot(), to execute a command, either within overlayroot-chroot or directly
source /opt/shift/scripts/include/exec_overlayroot.sh.inc

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# include function generateConfig() to generate config files from templates
source /opt/shift/scripts/include/generateConfig.sh.inc

# include errorExit() function
source /opt/shift/scripts/include/errorExit.sh.inc

# include updateTorOnions() function
source /opt/shift/scripts/include/updateTorOnions.sh.inc

# ------------------------------------------------------------------------------

redis_require

# update hardcoded Base image version
VERSION=$(head -n1 /opt/shift/config/version)
redis_set "base:version" "${VERSION}"


# check for reset triggers on flashdrive
if FLASHDRIVE="$(/opt/shift/scripts/bbb-cmd.sh flashdrive check)"; then
    echo "RESET: device ${FLASHDRIVE} detected"
    if /opt/shift/scripts/bbb-cmd.sh flashdrive mount "${FLASHDRIVE}"; then
        echo "RESET: flashdrive mounted"

        # are all necessary files for a reset present?
        if [[ -f /mnt/backup/.reset-token ]] && [[ -f /data/reset-token-hashes ]]; then
            FLASHDRIVE_TOKEN_HASH="$(sha256sum /mnt/backup/.reset-token | cut -f 1 -d " ")"
            echo "RESET: reset token present on flashdrive, hashed value: ${FLASHDRIVE_TOKEN_HASH}"

            # is hashed reset token present on Base?
            if grep -q "${FLASHDRIVE_TOKEN_HASH}" /data/reset-token-hashes; then
                echo "RESET: valid reset token found"

                if [[ -f /mnt/backup/reset-base-auth ]]; then
                    echo "RESET: trigger file 'reset-base-auth' found, initiating reset of authentication"
                    /opt/shift/scripts/bbb-cmd.sh reset auth --assume-yes
                    mv /mnt/backup/reset-base-auth /mnt/backup/reset-base-auth.done
                fi

                if [[ -f /mnt/backup/reset-base-config ]]; then
                    echo "RESET: trigger file 'reset-base-config' found. Feature not implemented yet."
                    mv /mnt/backup/reset-base-config /mnt/backup/reset-base-config.done
                fi

                if [[ -f /mnt/backup/reset-base-ssd ]]; then
                    echo "RESET: trigger file 'reset-base-ssd' found. Feature not implemented yet."
                    mv /mnt/backup/reset-base-ssd /mnt/backup/reset-base-ssd.done
                fi

                if [[ -f /mnt/backup/reset-base-image ]]; then
                    echo "RESET: trigger file 'reset-base-image' found. Feature not implemented yet."
                    mv /mnt/backup/reset-base-image /mnt/backup/reset-base-image.done
                fi
            else
                echo "RESET: reset token on flashdrive does not match authorized tokens on the Base"
            fi
        else
            echo "RESET: not all files for a reset present, doing nothing."
        fi

        umount /mnt/backup

    else
        echo "RESET: warning, could not mount flashdrive ${FLASHDRIVE}"
    fi
else
    echo "RESET: no flashdrive detected."
fi


# Base image updates
# ------------------------------------------------------------------------------
# initialize mender configuration
if [[ -f /etc/mender/mender.conf ]] && ! grep -q '/shift/' /etc/mender/mender.conf ; then
    exec_overlayroot all-layers 'rm -f /etc/mender/mender.* /etc/mender/server.crt || true'
    generateConfig mender.conf.template # -->  /etc/mender/mender.conf
fi


# update onion addresses in Redis
updateTorOnions


# check if rpcauth credentials exist, or create new ones
RPCAUTH="$(redis_get 'bitcoind:rpcauth')"
REFRESH_RPCAUTH="$(redis_get 'bitcoind:refresh-rpcauth')"

if [ ${#RPCAUTH} -lt 90 ] || [ "${REFRESH_RPCAUTH}" -eq 1 ] || [ "${REFRESH_RPCAUTH}" -eq -1 ]; then
    echo "INFO: creating new bitcoind rpc credentials"
    echo "INFO: old bitcoind:rpcauth was ${RPCAUTH}"
    echo "INFO: bitcoind:refresh-rpcauth is ${REFRESH_RPCAUTH}"
    /opt/shift/scripts/bbb-cmd.sh bitcoind refresh_rpcauth
else
    echo "INFO: found bitcoind rpc credentials, no action taken"
fi

# make sure Bitcoin-related services are enabled and started if setup is finished
if [[ $(redis_get "base:setup") -eq 1 ]] && ! systemctl is-enabled bitcoind.service; then
    echo "WARN: setup is completed, but Bitcoin-related services are not enabled. Enabling and starting them now..."
    /opt/shift/scripts/bbb-config.sh enable bitcoin_services
    /opt/shift/scripts/bbb-systemctl.sh start-bitcoin-services
fi
