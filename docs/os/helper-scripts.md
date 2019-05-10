---
layout: default
title: Helper scripts
parent: Operating System
nav_order: 250

---
## Helper scripts

Mostly for convenience during development, there are several helper scripts included. All scripts are located in the [`/opt/shift/scripts/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/scripts) directory.

* `set-bitcoin-network.sh`  
  Switch Bitcoin network for all services and configuration files.  
  ```
  BitBox Base: set Bitcoin network
  Usage: set-bitcoin-network.sh <testnet|mainnet>
  ```

* `systemd-base.sh`  
  Batch control all systemd units at once, e.g. for getting an overall status or stop all services.
  ```
  BitBox Base: batch control system units
  Usage: systemd-base.sh <status|start|restart|stop|enable|disable>
  ```
