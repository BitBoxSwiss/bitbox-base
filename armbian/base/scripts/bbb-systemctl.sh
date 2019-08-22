#!/bin/bash
set -eu

# BitBox Base: batch control system units
# 

function usage() {
    echo "BitBox Base: batch control system units"
    echo "Usage: bbb-systemctl <status|start|restart|stop|enable|disable>"
}

ACTION=${1:-"status"}

if [[ ${ACTION} == "-h" ]] || [[ ${ACTION} == "--help" ]]; then
  usage
  exit 0
fi

if ! [[ ${ACTION} =~ ^(status|start|restart|stop|enable|disable)$ ]]; then
  echo "bbb-systemctl.sh: unknown argument."
  echo
  usage
  exit 1
fi

case ${ACTION} in
        status)
                echo "
Checking systemd unit status of BitBox Base...

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
" | grep --color -E 'failed|activating|$'
                ;;

  start|restart|stop|enable|disable)

                if [[ ${UID} -ne 0 ]]; then
                  echo "bbb-systemctl.sh: needs to be run as superuser." >&2
                  exit 1
                fi

                systemctl daemon-reload

                systemctl $ACTION prometheus
                systemctl $ACTION prometheus-base.service 
                systemctl $ACTION prometheus-bitcoind.service 
                systemctl $ACTION prometheus-node-exporter.service 
                systemctl $ACTION grafana-server.service 
                systemctl $ACTION bbbmiddleware.service
                systemctl $ACTION nginx.service 
                systemctl $ACTION electrs.service 
                systemctl $ACTION lightningd.service 
                systemctl $ACTION bitcoind.service
                systemctl $ACTION bbbsupervisor.service
                systemctl $ACTION bbbfancontrol.service
                systemctl $ACTION redis.service
                ;;
esac
echo