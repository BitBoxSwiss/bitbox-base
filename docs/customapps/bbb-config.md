---
layout: default
title: bbb-config.sh
parent: Custom applications
nav_order: 140

---
## bbb-config.sh: system configuration utility

General system configuration utility, centrally defining recurring configuration actions.
This script can be used manually from the command line, but is also used by the BitBox Middleware and Supervisor to trigger configuration operations.

```
BitBoxBase: system configuration utility
usage: bbb-config.sh [--version] [--help]
                    <command> [<args>]

possible commands:
  enable    <bitcoin_incoming|bitcoin_ibd|bitcoin_ibd_clearnet|dashboard_hdmi|
             dashboard_web|wifi|autosetup_ssd|tor|tor_bbbmiddleware|tor_ssh|
             tor_electrum|overlayroot|sshpwlogin|rootlogin|unsigned_updates>

  disable   any 'enable' argument

  set       <hostname|loginpw|wifi_ssid|wifi_pw>
            bitcoin_network         <mainnet|testnet>
            bitcoin_dbcache         int (MB)
            other arguments         string
```

* **`enable`** or **`disable`** the following options:

  * `bitcoin_incoming`: Bitcoin Core accepts incoming connections
  * `bitcoin_ibd`: system is configured for Bitcoin initial block download (IBD) mode, setting `dbcache` to 2 GB and disabling `lightningd.service` and `electrs.service`.
  * `bitcoin_ibd_clearnet`: Bitcoin Core uses public internet for IBD and then automatically switch to Tor
  * `dashboard_hdmi`: Grafana system monitoring dashboard on the HDMI output. This feature is experimental and currently not available in official images, as the option BASE_HDMI_BUILD must also be enabled in [build.conf](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build.conf) when building the Armbian image.
  * `dashboard_web`: Grafana system monitoring dashboard available as website, e.g. under http://bitbox-base.local. For development only, off by default.
  * `wifi`: support for Wifi (experimental)
  * `autosetup_ssd`: detect an empty drive (ssd or hd, USB or PCIe) on boot and set it up automatically. See script [`autosetup-ssd.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/autosetup-ssd.sh) for details.
  * `tor`: Tor service, used for Bitcoin Core and c-lightning.
  * `tor_bbbmiddleware`: Tor hidden service for the BitBoxBase middleware, to connect from BitBoxApp, needs `tor` enabled
  * `tor_ssh`: Tor hidden service for SSH login, needs `tor` enabled
  * `tor_electrum`: Tor hidden service for Electrs, needs `tor` enabled
  * `overlayroot`: using `overlayrootfs`, mounting the root filesystem as read-only, with a ephemeral tmpfs overlay. Needs restart.
  * Insecure development options (see section [Tinkering](../tinkering)):
    * `sshpwlogin`: SSH login with password
    * `rootlogin`: SSH login as root user
    * `unsigned_updates`: allow updating from unsigned Mender images

* **set** the following option:

  * `hostname`: set hostname persistently
  * `loginpw`: change login/sudo password for both users `base` and `root`, will be overwritten when running the BitBoxApp Setup Wizard again
  * `wifi_ssid`: [experimental] SSID for wifi
  * `wifi_pw`: [experimental] PW for wifi
  * `bitcoin_network`: Bitcoin network, either `mainnet` or `testnet`
  * `bitcoin_dbcache`: set `dbcache` option for Bitcoin Core
