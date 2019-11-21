#!/usr/bin/env python3
# -*- coding: utf-8 -*-
#
# This script is called by the prometheus-base.service
# to provide system metrics to Prometheus.
#

import json
import time
import subprocess
import sys
import socket
import redis
from prometheus_client import start_http_server, Gauge, Counter, Info

# Create Prometheus metrics to track Base stats.
## metadata
BASE_SYSTEM_INFO = Info("base_system", "System information")

## System metrics
BASE_CPU_TEMP = Gauge("base_cpu_temp", "CPU temperature")
BASE_FAN_SPEED = Gauge("base_fan_speed", "Fan speed in %")

## systemd status: 0 = running / 3 = inactive
BASE_SYSTEMD_BITCOIND = Gauge("base_systemd_bitcoind", "Systemd unit status for Bitcoin Core")
BASE_SYSTEMD_ELECTRS = Gauge("base_systemd_electrs", "Systemd unit status for Electrs")
BASE_SYSTEMD_LIGHTNINGD = Gauge("base_systemd_lightningd", "Systemd unit status for c-lightning")
BASE_SYSTEMD_PROMETHEUS = Gauge("base_systemd_prometheus", "Systemd unit status for Prometheus")
BASE_SYSTEMD_GRAFANA = Gauge("base_systemd_grafana", "Systemd unit status for Grafana")

r = redis.Redis(
    host='127.0.0.1',
    port=6379,
)

def readFile(filepath):
    args = [filepath]
    try:
        with open(filepath) as f:
            value = f.readline()
    except:
            value = "err"

    return value

def getIP():
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        # doesn't even have to be reachable
        s.connect(('192.0.2.0', 1))
        IP = s.getsockname()[0]
    except:
        IP = '127.0.0.1'
    finally:
        s.close()
    return IP

def getSystemInfo():
    rediskeys = ['base:hostname','base:version','build:date','build:time','build:commit']
    info = {}
        # add Redis values
    for k in rediskeys:
        infoName = k.lower().replace(":", "_")
        infoValue = r.get(k)
        if infoValue is not None:
            infoValue = infoValue.decode("utf-8")
        else:
            infoValue = 'n/a'

        info[infoName] = infoValue

    info['base_ipaddress'] = getIP()
    return info

def getSystemdStatus(unit):
    try:
        subprocess.check_output(["systemctl", "is-active", unit])
        return 0
    except subprocess.CalledProcessError as e:
        print(unit, e.returncode, e.output)
        return e.returncode

def main():
    # Start up the server to expose the metrics.
    start_http_server(8400)
    while True:

        BASE_SYSTEM_INFO.info(getSystemInfo())
        BASE_SYSTEMD_BITCOIND.set(int(getSystemdStatus("bitcoind")))
        BASE_SYSTEMD_ELECTRS.set(int(getSystemdStatus("electrs")))
        BASE_SYSTEMD_LIGHTNINGD.set(int(getSystemdStatus("lightningd")))
        BASE_SYSTEMD_PROMETHEUS.set(int(getSystemdStatus("prometheus")))
        BASE_SYSTEMD_GRAFANA.set(int(getSystemdStatus("grafana-server")))

        try:
            BASE_CPU_TEMP.set(readFile("/sys/class/thermal/thermal_zone0/temp"))
        except:
            BASE_CPU_TEMP.set(0)

        try:
            BASE_FAN_SPEED.set(readFile("/sys/class/hwmon/hwmon0/pwm1"))
        except:
            BASE_FAN_SPEED.set(0)

        time.sleep(10)


if __name__ == "__main__":
    main()
