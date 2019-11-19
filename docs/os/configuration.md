---
layout: default
title: Configuration
parent: Operating system
nav_order: 200

---
## Configuration

It's important to keep the BitBoxBase in a consistently configured state.
Here we describe how to control the configuration during operations, how to store and backup the it and manage it remotely from the BitBoxApp.

### *bbb-config.sh*: manage configuration during operations

System configuration is managed internally using the script [`bbb-config.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-config.sh).
Its goal is to centrally define how changes are applied to the system and reuse a single set of commands.
This is why it is called by the build script as well as by the BitBoxBase Middleware during operations.
Changes are applied by simple operating system commands like copying and deleting files, or replacing text values withing configuration files.
It also takes care of writing the changes into the read-only rootfs when necessary.

The script can also be used directly from the command line, requiring sudo privileges. Call the script `bbb-config.sh --help` to see all possible commands and arguments:

```
BitBoxBase: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

assumes Redis database running to be used with 'redis-cli'

possible commands:
  enable    <bitcoin_services|bitcoin_incoming|bitcoin_ibd|bitcoin_ibd_clearnet|dashboard_hdmi|
             dashboard_web|wifi|autosetup_ssd|tor|tor_bbbmiddleware|tor_ssh|
             tor_electrum|overlayroot|sshpwlogin|rootlogin|unsigned_updates>

  disable   any 'enable' argument

  set       <hostname|loginpw|wifi_ssid|wifi_pw>
            bitcoin_network         <mainnet|testnet>
            bitcoin_dbcache         int (MB)
            other arguments         string
```

### *bbb-cmd.sh*: execution of standard commands

Similar to the configuration script, the [`bbb-cmd.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-cmd.sh) script acts as the central repository for standard commands, to be called from the Middleware. Or run it from the commandline with sudo.

```
BitBoxBase: system commands repository
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

### *Redis*: storage of configuration values

The [Redis](https://redis.io/) key/value datastore is used to manage configuration data.
It can be queried from all software components, be it from the command line, Bash, Python or Go with minimal overhead.

* from the terminal, Redis can be used with its command-line utility [`redis-cli`](https://redis.io/topics/rediscli)
* for usage within bash scripts, the necessary helper function are sourced from the include [redis.sh.inc](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/include/redis.sh.inc)
* Go applications use the [Redigo](https://github.com/gomodule/redigo) client
* Python uses [redis-py](https://github.com/andymccurdy/redis-py)

During build, the factory settings are imported from [`armbian/base/config/redis/factorysettings.txt`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/config/redis/factorysettings.txt) using the [Redis mass insertion protocol](https://redis.io/topics/mass-insert) using the helper script [`armbian/base/scripts/redis-pipe.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/redis-pipe.sh).

The Redis data is dumped frequently and on demand to `/data/redis/bitboxbase.rdb`, from where it can be backed up or restored.

Configuration values are stored in keys, like the following examples. For a full reference of used keys, refer to [`factorysettings.txt`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/config/redis/factorysettings.txt).

```
base:hostname         bitbox-base
tor:base:enabled      1
bitcoind:network      mainnet
lightningd:bind-addr  127.0.0.1:9735
electrs:db_dir        /mnt/ssd/electrs/db
```

### *bbbconfgen*: Updating of configuration files

While the Middleware and some scripts query Redis directly, most keys need to be written into configuration files to take effect.
Configuration files are created dynamically using [`bbbconfgen`](https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbconfgen), with the templates located in [`armbian/base/config/templates/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/config/templates):

This application parses a template, populates it with the corresponding Redis values, and stores it to the system (even into the read-only filesystem, if applicable).

* during build, the bash function `generateConfig()` is used within the Armbian customizing script.
* in regular operation, changes to Redis values and the regeneration of config files are typically executed through the `bbb-config.sh` or `bbb-cmd.sh` scripts.

To keep the configuration scripts consistent, the bash function `generateConfig()` is sourced from the include file [`generateConfig.sh.inc`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/include/generateConfig.sh.inc).

If `bbbconfgen` is run from the commandline and overlay root filesystem is enabled, you need to make sure that the configuration file is not only written into the tmpfs overlay. Either disable overlayrootfs (and reboot first), or use the application within `overlayroot-chroot`.

### *BitBoxApp*: user interface

Ultimately, the configuration is managed by the user through the BitBoxApp, that talks to the Middleware which in turn calls either the `bbb-config.sh` or `bbb-cmd.sh` script with the necessary arguments.

The App provides a convenient backup feature to save the whole system configuration directly to a USB stick plugged into the Base.

See the section [User Interface](../ui) for more details.
