---
layout: default
title: Prometheus
parent: Applications
nav_order: 150
---
## Prometheus: System Monitoring

There's a lot going on within the BitBoxBase and it's good to have reliable metrics over time to aid in development, fine-tuning, during operations and when looking for the cause of a specific error.
[Prometheus](https://prometheus.io/) takes care of that.
It is a open-source monitoring solution that is easily extensible to monitor any application.
Measurements are stored in a time-series database that can easily be queried and visualized, e.g. using [Grafana](grafana.md).
This database is also the main source of data that is queried by the Base Middleware.

### Installation

Prometheus is downloaded directly from the GitHub releases page and checked against a hardcoded checksum.
The `.tar.gz` archive is extracted, its content copied to various system locations and proper ownerwhip is set while building the Armbian image.

```bash
mkdir -p /etc/prometheus /var/lib/prometheus
cp prometheus promtool /usr/local/bin/
cp -r consoles/ console_libraries/ /etc/prometheus/
chown -R prometheus /etc/prometheus /var/lib/prometheus
```

### Configuration

Applications providing metrics basically just run a webserver that returns plain-text content in a specific format on a certain port.
Prometheus queries these URI in fixed intervals and stores the metrics in its database.
Where to get this information how often is specified in `/etc/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval:     1m
  evaluation_interval: 1m
scrape_configs:
  - job_name: node
    static_configs:
      - targets: ['127.0.0.1:9100']
  - job_name: base
    static_configs:
      - targets: ['127.0.0.1:8400']
  - job_name: bitcoind
    static_configs:
      - targets: ['127.0.0.1:8334']
  - job_name: electrs
    static_configs:
    - targets: ['127.0.0.1:4224']
  - job_name: lightningd
    static_configs:
    - targets: ['127.0.0.1:9900']
```

### Metrics

The following metrics are collected both from the system and from specific applications.

* **Operating System**: the [Prometheus Node Exporter](https://github.com/prometheus/node_exporter) collects very granular information about the system, such as CPU and memory usage.
  * Installation: downloaded from the [GitHub releases page](https://github.com/prometheus/node_exporter/releases), verified against a hardcoded checksum and installed by the Armbian build script similar to Prometheus itself
  * Service management: run by systemd as `prometheus-node-exporter.service`
  * Prometheus URI: <http://127.0.0.1:9100>
* **BitBoxBase**
  * Installation: single Python3 script [`prometheus-base.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/prometheus-base.py) that is copied to `/opt/shift/scripts/`
  * Service management: run by systemd as `prometheus-base.service`
  * Prometheus URI: <http://127.0.0.1:8400>
* **Bitcoin Core**
  * Installation: single Python3 script [`prometheus-bitcoind.py`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/prometheus-bitcoind.py) that is copied to `/opt/shift/scripts/`
  * Service management: run by systemd as `prometheus-bitcoind.service`
  * Prometheus URI: <http://127.0.0.1:8334>
* **electrs**
  * Installation & service management: `electrs` provides Prometheus metrics by default, no additional steps necessary.
  * Prometheus URI: <http://127.0.0.1:4224>
* **c-lightning**
  * Installation: single Python3 script that is downloaded from the [`lightningd/plugins`](https://github.com/lightningd/plugins/tree/master/prometheus) GitHub repository to `/opt/shift/scripts/` and checked against a hardcoded checksum
  * Service management: the script is run as a c-lightning server plugin and started together with `lightningd`.
    It is specified in the configuration file `/etc/lightningd/lightningd.conf`.
  * Prometheus URI: <http://127.0.0.1:9900>

### Service management

The Prometheus service itself is managed by systemd.
Relevant parameters are specified in the unit file `/etc/systemd/system/prometheus.service` shown below.

```
[Unit]
Description=Prometheus
After=network-online.target

[Service]
User=prometheus
Group=system
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --web.listen-address="127.0.0.1:9090" \
    --config.file /etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/mnt/ssd/prometheus \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Data storage

The database is stored on the SSD in the `/mnt/ssd/prometheus/` directory.
