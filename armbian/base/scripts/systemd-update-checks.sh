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

# update onion addresses in Redis
updateTorOnions

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

# check if booting after update
# valid status codes of 'base:updating'
#    0: no update in progress
#   10: mender update applied
#   20: system reconfigured
#   30: system tested OK
#   40: mender update committed
#   90: mender update failed

# if Redis key is not set, assume update
if [[ $(redis_get "base:updating") == "" ]]; then
    echo "INFO: Redis key 'base:updating' is not set, assuming update"
    redis_set "base:updating" 10
fi

if [[ $(redis_get "base:updating") -eq 0 ]]; then
    echo "INFO: not updating"

else
    # after first boot into new rootfs, update configuration files
    if [[ $(redis_get "base:updating") -eq 10 ]]; then

        /opt/shift/scripts/bbb-config.sh set hostname "$(redis_get 'base:hostname')"

        generateConfig "bashrc-custom.template"
        generateConfig "bbbmiddleware.conf.template"
        generateConfig "bitboxbase.service.template"
        generateConfig "bitcoin.conf.template"
        generateConfig "electrs.conf.template"
        generateConfig "grafana.ini.template"
        generateConfig "mender.conf.template"
        generateConfig "iptables.rules.template"
        generateConfig "lightningd.conf.template"
        generateConfig "torrc.template"

        if [[ $(redis_get "base:wifi:enabled") -eq 1 ]]; then
            generateConfig "wlan0.conf.template"
            echo "INFO: restarted Tor and networking daemon"
        fi
        echo "INFO: system configuration recreated"

        # restart services (don't exit on failure)
        if [[ $(redis_get "tor:base:enabled") -eq 1 ]]; then
            systemctl restart tor.service || true
            echo "INFO: restarted Tor and networking daemon"
        fi

        set -x
        systemctl restart iptables-restore          || true
        systemctl restart avahi-daemon.service      || true
        systemctl restart networking.service        || true
        systemctl restart bbbmiddleware.service     || true
        systemctl restart bitcoind.service          || true
        systemctl restart electrs.service           || true
        systemctl restart lightningd.service        || true
        systemctl restart grafana-server.service    || true

        set +x

        echo "OK: restarted all reconfigured services"
        redis_set "base:updating" 20
    fi

    # validate that all services are running
    if [[ $(redis_get "base:updating") -eq 20 ]]; then
        for run in {1..100}; do
            if  /opt/shift/scripts/bbb-systemctl.sh verify; then
                redis_set "base:updating" 30
                break
            fi

            # if verification unsuccessful, try again or reboot to fall back
            if [[ ${run} -lt 100 ]]; then
                echo "INFO: service verification try ${run} of 100 unsuccessful, retrying in 10 seconds"
                sleep 2
            else
                echo "ERR: service verification try ${run} of 100 unsuccessful, falling back to previous Base image version"
                redis_set "base:updating" 90
                /opt/shift/scripts/bbb-cmd.sh base restart
                errorExit MENDER_UPDATE_SYSCONFIG_FAILED
            fi
        done

    fi

    if [[ $(redis_get "base:updating") -eq 30 ]]; then
        if /opt/shift/scripts/bbb-cmd.sh mender-update commit; then
            echo "OK: mender commit ${?}"
            redis_set "base:updating" 40
        else
            echo "ERR: mender commit failed with error code ${?}"
        fi
    fi

    if [[ $(redis_get "base:updating") -eq 40 ]]; then
        echo "OK: updated to BitBoxBase version $(redis_get 'base:version')"
        redis_set "base:updating" 0
    fi

    if [[ $(redis_get "base:updating") -ne 0 ]]; then
        echo "ERR: undefined value $(redis_get 'base:updating') for Redis key 'base:updating'"
    fi

fi
