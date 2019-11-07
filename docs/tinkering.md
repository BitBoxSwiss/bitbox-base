---
layout: default
title: Tinkering
nav_order: 800
---
# BitBoxBase: Tinkering

For **advanced users**, full root access and the possibility to validate and customize the BitBoxBase configuration is a must. This section explains how to work with the device on a operating-system level. Be aware that you could easily break stuff and might not be able to recover the factory settings easily.

## SSH access

SSH access is disabled by default. Once enabled, you should always log in with the user `base` that has sudo privileges. The user password corresponds to the one set in the BitBoxApp setup wizard.

There are multiple ways to gain access, some usable for production, others only suitable to be used for development:

* **SSH keys**: if SSH keys are present in `/home/base/.ssh/authorized_keys`, SSH login is possible over regular IP address, the mDNS domain (e.g. `ssh base@bitbox-base.local`) or even a Tor hidden service (if enabled).

  Currently, the keys need to be added manually, either by logging in locally or after login in with a password (see next option). We plan to allow users to add SSH keys from the BitBoxApp.
* **Password login**: this authentication method is not secure and should not be enabled for longer periods on a production device.

  It can be enabled in the BitBoxApp node management under "Advanced options". Alternatively, you can run `sudo bbb-config.sh enable sshpwlogin` directly on the command line. After enabling, you can log in with the user `base` using the password set in the Setup wizard.
* **Root login**: SSH access for the `root` user is disabled by default. For development, it can be enabled from the command line, e.g. to copy updated scripts directly into system folders that require root access. On the BitBoxBase, logged in with user `base`, run `sudo bbb-config.sh enable rootlogin`.

If you build the BitBoxBase image yourself, you can configure the options `BASE_LOGINPW` (initial login password, overwritten by the Setup Wizard) and `BASE_SSH_PASSWORD_LOGIN` in [build.conf](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build.conf).

## Disable overlayroot filesystem

You must be aware that the whole **root filesystem is mounted as read-only**, and a temporary in-memory filesystem is used as an overlay. That means that you can change anything you want on the rootfs of the BitBoxBase (not the SSD, though!), but **on reboot, all changes are gone**.

This is great for resilience, eg. no corruption can occur on a sudden power outage, but it's not great for tinkering with your device.

For quick changes to the read-only filesystem, you can chroot into it:

```
$ sudo overlayroot-chroot
INFO: Chrooting into [/media/root-ro]
...
# do your stuff
...
$ exit
```

To deactivate the overlay root filesystem, run `sudo bbb-config.sh disable overlayrootfs` from the command line. After a reboot, the rootfs is mounted regularly and all changes are persistent. You can always `enable` the overlayrootfs again, using the same script. Each official update overwrites the whole rootfs, however.

If you build the BitBoxBase image yourself, you can configure the option `BASE_OVERLAYROOT` in [build.conf](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build.conf).

## Configuration: bbb-config.sh

Most configuration options can be enabled, disabled or set using the script [`bbb-config.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-config.sh). It must be run with sudo privileges and is the central repository to make persistent configuration changes, as it also writes into the read-only rootfs when necessary.

```
$ bbb-config.sh

BitBox Base: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

assumes Redis database running to be used with 'redis-cli'

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

See section [Custom applications / Helper scripts](helper-scripts.md#bbb-configsh-configuration-management) for more information.

## Commands: bbb-cmd.sh

Recurring commands are available in the central repository [`bbb-cmd.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-cmd.sh). Run it with sudo directly from the command line.

```
$ bbb-cmd.sh

BitBox Base: system commands repository
usage: bbb-cmd.sh [--version] [--help] <command>

possible commands:
  setup         <datadir>
  bitcoind      <reindex|resync|refresh_rpcauth>
  flashdrive    <check|mount|umount>
  backup        <sysconfig|hsm_secret>
  restore       <sysconfig|hsm_secret>
  reset         <auth|config|image|ssd>
  mender-update <install|commit>
```

See section [Custom applications / Helper scripts](helper-scripts.md#bbb-cmdsh-execution-of-standard-commands) for more information.

## Redis configuration management

The scripts above manipulate the configuration management database that is provided by [Redis](applications/redis.md). You can manually read and set the value of Redis keys using its command line application:

```
$ redis-cli get bitcoind:txindex
"0"
$ redis-cli set bitcoind:txindex 1
OK
```

All keys that are used are listed in the [Redis factory settings](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/config/redis/factorysettings.txt). Keys that are populated on first boot are listed with the value `xxx`.

## Configuration templates

While the Middleware and some scripts query Redis directly, most keys need to be written into configuration files to take effect. All files that contain changing values are available as templates located in [`/opt/shift/config/templates/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/config/templates) and can be regenerated with the [BitBoxBase Confgen](customapps/bbbconfgen.md).

In the above example, the Bitcoin Core transaction index is configured, but not yet written to the file `/etc/bitcoin/bitcoin.conf`. The file [`bitcoin.conf.template`] contains references to the target location and all relevant Redis keys. To recreate the configuration file, run the following command:

```
$ bbbconfgen --template /opt/shift/config/templates/bitcoin.conf.template

connected to Redis
opened template config file /opt/shift/config/templates/bitcoin.conf.template
writing into output file /etc/bitcoin/bitcoin.conf
written 28 lines
placeholders: 19 replaced, 0 kept, 0 deleted, 0 lines deleted, 1 set to default
checks: 4 lines dropped, 4 lines kept
```

If the overlay root filesystem is enabled, you need to make sure that the configuration file is not only written into the tmpfs overlay. Either disable overlayrootfs (and reboot first), or use `overlayroot-chroot`.

## Factory reset

It is possible to reset various aspects the BitBoxBase to factory settings, either through the BitBoxApp (not implemented yet) or by creating a trigger file on the backup USB flashdrive.

In case of a forgotten password (set in the BitBoxApp Setup Wizard), it can be reset as follows:

* On the backup flashdrive, create a file named `reset-base-auth`.
* The flashdrive must contain a valid reset token, created on initial setup or subsequent backups
* Plug the flashdrive into the BitBoxBase
* Restart the device
* The authentication is now reset, run the BitBoxApp Setup Wizard again to set a new password.
