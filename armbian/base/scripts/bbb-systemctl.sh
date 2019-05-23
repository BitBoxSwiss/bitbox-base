#!/bin/bash
set -eu

# BitBox Base: batch control system units
# 

function usage() {
    echo "BitBox Base: batch control system units"
    echo "Usage: bbb-systemctl <status|start|restart|stop>"
}

ACTION=${1:-"status"}

if ! [[ ${ACTION} =~ ^(status|start|restart|stop|enable|disable)$ ]]; then
	usage
	exit 1
fi

if [[ ${UID} -ne 0 ]]; then
  echo "${0}: needs to be run as superuser." >&2
  exit 1
fi

echo 
case ${ACTION} in
        status)
                echo "Checking systemd unit status of BitBox Base..."
                echo 
                echo "bitcoind:                 $(systemctl is-active bitcoind.service)"
                echo "electrs:                  $(systemctl is-active electrs.service)"
                echo "lightningd:               $(systemctl is-active lightningd.service)"
                echo "base-middleware:          $(systemctl is-active base-middleware.service)"
                echo "nginx:                    $(systemctl is-active nginx.service)"
                echo "prometheus:               $(systemctl is-active prometheus.service)"
                echo "prometheus-node-exporter: $(systemctl is-active prometheus-node-exporter.service)"
                echo "prometheus-base:          $(systemctl is-active prometheus-base.service)"
                echo "prometheus-bitcoind:      $(systemctl is-active prometheus-bitcoind.service)"
                echo "grafana:                  $(systemctl is-active grafana-server.service)"
                echo "bbbfancontrol:            $(systemctl is-active bbbfancontrol.service)"
                echo
                ;;

	start|restart|stop|enable|disable)
                systemctl daemon-reload

                systemctl $ACTION prometheus
                systemctl $ACTION prometheus-base.service 
                systemctl $ACTION prometheus-bitcoind.service 
                systemctl $ACTION prometheus-node-exporter.service 
                systemctl $ACTION grafana-server.service 
                systemctl $ACTION base-middleware.service
                systemctl $ACTION nginx.service 
                systemctl $ACTION electrs.service 
                systemctl $ACTION lightningd.service 
                systemctl $ACTION bitcoind.service
                systemctl $ACTION bbbfancontrol.service
                ;;
esac
echo