---
layout: default
title: Helper scripts
parent: Operating System
nav_order: 250

---
## Helper scripts

Mostly for convenience during development, there are several helper scripts included. All scripts are located in the [`/opt/shift/scripts/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/scripts) directory.

* `bbb-config.sh`  
  General system configuration utility. 
  Settings are stored in individual files in `/opt/shift/sysconfig/` as key/value pair. 
  For example, the file `BITCOIN_NETWORK` contains `BITCOIN_NETWORK=mainnet`. 
  It can be sourced by any script, so that the variable `BITCOIN_NETWORK` is available immediately.
  ```
  BitBox Base: system configuration utility
  usage: bbb-config [--version] [--help]
                  <command> [<args>]

  possible commands:
  enable    <dashboard_hdmi|dashboard_web|wifi|autosetup_ssd>

  disable   any 'enable' argument

  set       <bitcoin_network|hostname|root_pw|wifi_ssid|wifi_pw>
            bitcoin_network     <mainnet|testnet>
            other arguments     string

  get       any 'enable' or 'set' argument, or <all>
  ```

* `bbb-systemctl.sh`  
  Batch control all systemd units at once, e.g. for getting an overall status or stop all services.
  ```
  BitBox Base: batch control system units
  Usage: bbb-systemctl.sh <status|start|restart|stop|enable|disable>
  ```
