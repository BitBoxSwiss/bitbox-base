---
layout: default
title: Additional scripts
parent: Custom applications
nav_order: 160

---
## Additional scripts

Several helper scripts are located in the [`/opt/shift/scripts/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/scripts) directory.

### Interactive use

* [`bbb-systemctl.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-systemctl.sh): manage and check systemd units in batch
  Batch control all systemd units at once, e.g. for getting an overall status or stop all services.

  ```
  BitBoxBase: batch control system units
  Usage: bbb-systemctl.sh <status|start|restart|stop|enable|disable|verify>
  ```

### Startup

* [`systemd-startup-checks.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/systemd-startup-checks.sh): early system checks
* [`systemd-startup-after-redis.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/systemd-startup-after-redis.sh): second stage system checks, after Redis is available
* [`systemd-update-checks.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/systemd-update-checks.sh): checks if update is in progress and performs system checks, update commits or fallback


### Systemd startpre / startpost

Scripts named `systemd-***-startpre` and `systemd-***-startpost` are executed by systemd before or after starting system services.

### Various

* [`bitcoind-rpcauth.py`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bitcoind-rpcauth.py): generate Bitcoin Core RPC authentication
* [`prometheus-base.py`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/prometheus-base.py): Prometheus scraper for custom BitBoxBase components
* [`prometheus-bitcoind.py`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/prometheus-bitcoind.py): Prometheus scraper for Bitcoin Core
* [`prometheus-lightningd.py`](https://github.com/lightningd/plugins/tree/master/prometheus): Prometheus scraper by [Christian Decker](https://github.com/cdecker) for c-lightning
* [`redis-pipe.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/redis-pipe.sh): script for mass insertion of Redis key/value data
