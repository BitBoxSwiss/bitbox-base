---
layout: default
title: Helper scripts
parent: Custom applications
nav_order: 140

---
## Helper scripts

Several helper scripts are located in the [`/opt/shift/scripts/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/scripts) directory.


### [**bbb-config.sh**](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-config.sh): configuration management
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

### [**bbb-cmd.sh**](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-cmd.sh): execution of standard commands

Similar to the configuration script, the [`bbb-cmd.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-cmd.sh) script acts as the central repository for standard commands, mainly to be called from the Middleware.

```
BitBoxBase: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  setup         <datadir>
  bitcoind      <reindex|resync|refresh_rpcauth>
  flashdrive    <check|mount|umount>
  backup        <sysconfig|hsm_secret>
  restore       <sysconfig|hsm_secret>
  mender-update <install|commit>
```

The following commands are available:

* **setup**:
  * **datadir**: called on boot to check if the persistent data directory is already set up, and initializes it if necessary. Multiple scenarios need to be considered:
    * read/write disk, without Mender (likely a development image)
      the `/data` directory can be created directly, copying the initial content from the source directory `/data_source`
    * read-only disk, without Mender (e.g. a DIY image)
      image is build using `OVERLAYROOT` option, so all changes written to the root filesystem are volatile due to the tmpfs overlay. To preserve changes in `/data`, the directory is created as a symbolic link to `/mnt/ssd/data` on the SSD
    * read-only disk, with Mender (BitBoxBase production image)
      the `/data` directory is mounted from a separate, persistent partition that is not overwritten on update. Initial content is copied once from `/data_src` into that partition.

* **bitcoind**
  * **reindex**: deletes the Bitcoin Core chainstate (UTXO set) and the Electrs indices, but not the raw blockchain data. Bitcoin Core is restarted to reindex the whole existing blockchain, thus building up a new UTXO set and validating the whole blockchain from Genesis.
  * **resync**: in addition to *reindex*, this command also deletes the raw blockchain data. After restarting Bitcoin Core, the whole blockchain data (~250 GB) are downloaded before a full validation is conducted.
  * **refresh_rpcauth**: authentication to Bitcoin Core JSON API uses the `rpcauth` method, with clients using static `rpcuser` and `rpcpassword` values. This command automatically creates new authentication keys and recreates related application configuration files.

* **flashdrive**: controls a USB flashdrive plugged directly in to the device
  * **check**: checks if a USB flashdrive suitable for a backup is plugged in, and returns its device path (e.g. `/dev/sdb1`). Exactly one flashdrive must be present that is not larger than 64 GiB, so exteral USB drives in DIY builds are not confused with flashdrives.
  * **mount**: expects the device path as an argument and mounts the flashdrive to `/mnt/backup` with a restrictive usage policy if it is suited for backup
  * **unmount**: unmounts the flashdrive

* **backup**
  * **sysconfig**: copies the Redis datastore `/data/redis/bitboxbase.rdb` to a mounted flashdrive
  * **hsm_secret**: stores the c-lightning on-chain seed `/mnt/ssd/bitcoin/.lightning/hsm_secret` into Redis, encoded as base64

* **restore**
  * **sysconfig**: restores the Redis datastore `/data/redis/bitboxbase.rdb` from a mounted flashdrive and restarts Redis
  * **hsm_secret**: saves the c-lightning on-chain seed from Redis into `/mnt/ssd/bitcoin/.lightning/hsm_secret`

* **mender-update**
  * **install**: expects a Base image version and downloads/verifies/installs the Mender update artefact into the inactive partition. A reboot is required to boot into the updated system. This command requireds either a version number (in the format of `x.x.x`) to fetch the release from GitHub, or `flashdrive` to load the update from a mounted usb flashdrive (update image must be available as `/mnt/backup/update.base`)
  * **commit**: commits an update to become persistent. If it is not committed, the device falls back to the previous Base image on reboot.

### [**bbb-systemctl.sh**](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-systemctl.sh): manage and check systemd units in batch
Batch control all systemd units at once, e.g. for getting an overall status or stop all services.
```
BitBoxBase: batch control system units
Usage: bbb-systemctl.sh <status|start|restart|stop|enable|disable|verify>
```
