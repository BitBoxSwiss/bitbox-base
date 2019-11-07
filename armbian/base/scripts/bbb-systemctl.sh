#!/bin/bash
# shellcheck disable=SC1091
set -eu

# BitBoxBase: batch control system units
#

# --- generic functions --------------------------------------------------------

# include functions redis_set() and redis_get()
source /opt/shift/scripts/include/redis.sh.inc

# include errorExit() function
source /opt/shift/scripts/include/errorExit.sh.inc

# ------------------------------------------------------------------------------

function usage() {
    echo "BitBoxBase: batch control system units"
    echo "Usage: bbb-systemctl <status|start|restart|stop|enable|disable>"
}

ACTION=${1:-"status"}

if [[ ${ACTION} == "-h" ]] || [[ ${ACTION} == "--help" ]]; then
    usage
    exit 0
fi

if ! [[ ${ACTION} =~ ^(status|start|restart|stop|enable|disable|verify)$ ]]; then
    echo "bbb-systemctl.sh: unknown argument."
    echo
    usage
    exit 1
fi

case ${ACTION} in
    status)
        echo "
Checking systemd unit status of BitBoxBase...
highlight: failed, activating, inactive

bitcoind:                 $(systemctl is-active bitcoind.service)
electrs:                  $(systemctl is-active electrs.service)
lightningd:               $(systemctl is-active lightningd.service)
bbbmiddleware:            $(systemctl is-active bbbmiddleware.service)
bbbsupervisor:            $(systemctl is-active bbbsupervisor.service)
bbbfancontrol:            $(systemctl is-active bbbfancontrol.service)
redis:                    $(systemctl is-active redis.service)
nginx:                    $(systemctl is-active nginx.service)
prometheus:               $(systemctl is-active prometheus.service)
prometheus-node-exporter: $(systemctl is-active prometheus-node-exporter.service)
prometheus-base:          $(systemctl is-active prometheus-base.service)
prometheus-bitcoind:      $(systemctl is-active prometheus-bitcoind.service)
grafana:                  $(systemctl is-active grafana-server.service)
" | grep --color -zP  '(failed|activating|inactive)'
        ;;

    start|restart|stop|enable|disable)

        if [[ ${UID} -ne 0 ]]; then
            echo "bbb-systemctl.sh: needs to be run as superuser." >&2
            exit 1
        fi

        systemctl daemon-reload

        systemctl "$ACTION" prometheus
        systemctl "$ACTION" prometheus-base.service
        systemctl "$ACTION" prometheus-bitcoind.service
        systemctl "$ACTION" prometheus-node-exporter.service
        systemctl "$ACTION" grafana-server.service
        systemctl "$ACTION" bbbmiddleware.service
        systemctl "$ACTION" nginx.service
        systemctl "$ACTION" electrs.service
        systemctl "$ACTION" lightningd.service
        systemctl "$ACTION" bitcoind.service
        systemctl "$ACTION" bbbsupervisor.service
        systemctl "$ACTION" bbbfancontrol.service
        systemctl "$ACTION" redis.service
        ;;

    verify)
        # verify that all services are running

        if  systemctl is-active -q prometheus                       && \
            systemctl is-active -q prometheus-base.service          && \
            systemctl is-active -q prometheus-bitcoind.service      && \
            systemctl is-active -q prometheus-node-exporter.service && \
            systemctl is-active -q grafana-server.service           && \
            systemctl is-active -q bbbmiddleware.service            && \
            systemctl is-active -q nginx.service                    && \
            systemctl is-active -q bitcoind.service                 && \
            systemctl is-active -q bbbsupervisor.service            && \
            systemctl is-active -q bbbfancontrol.service            && \
            systemctl is-active -q redis.service
        then
            if  [[ "$(redis_get 'bitcoind:ibd')" -ne 1 ]]; then
                if  ! systemctl is-active -q electrs.service        || \
                    ! systemctl is-active -q lightningd.service
                then
                    echo "ERR: bitcoind not in IBD mode, but lightningd and/or electrs not running"
                    errorExit SYSTEMD_NOT_ALL_SERVICES_RUNNING
                fi

            else
                echo "OK: bitcoind is in IBD mode"
            fi

            echo "OK: all services are active"
            exit 0

        else
            echo "ERR: not all services are active"
            errorExit SYSTEMD_NOT_ALL_SERVICES_RUNNING
        fi
        ;;
esac
