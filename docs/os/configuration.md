---
layout: default
title: Configuration
parent: Operating System
nav_order: 320

---
## Configuration

It's important to keep the BitBox Base in a consistently configured state.
Here we describe how to set the initial configuration on build, control it internally during operations, store and backup the configuration and manage it remotely from the BitBox App.

### Initial configuration on build

The initial system configuration is set on build and can be altered by setting build options in the file [`armbian/base/build/build.conf`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/build.conf).  

Available options are described directly in the file and are set to default values. 
A few examples of build options you can set:
* `BASE_BITCOIN_NETWORK`: set to `mainnet` or `testnet`
* `BASE_HOSTNAME`: set it to `alice` and your BitBox Base will be visible as `alice.local` within your network
* `BASE_WIFI_SSID` and `BASE_WIFI_PW`: configure the image to connect to your wireless network
* ...and many more.

To preserve a local configuration, you can copy the file to `build-local.conf` in the same directory. This file is excluded from Git source control and overwrites options from `build.conf`.

### Manage configuration during operations
System configuration is managed internally using the script [`bbb-config.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-config.sh). 
Its goal is to centrally define how changes are applied to the system and reuse a single set of commands.
This is why it is called by the build script as well as by the BitBox Base Middleware during operations.
Changes are applied by simple operating system commands like copying and deleting files, or replacing text values withing configuration files.

You can call the script `bbb-config.sh --help` to see all possible commands and arguments:

```
BitBox Base: system configuration utility
usage: bbb-config.sh [--version] [--help]
                     <command> [<args>]

possible commands:
enable    <dashboard_hdmi|dashboard_web|wifi|autosetup_ssd|
           tor_ssh|tor_electrum>

disable   any 'enable' argument

set       <bitcoin_network|hostname|root_pw|wifi_ssid|wifi_pw>
          bitcoin_network     <mainnet|testnet>
          other arguments     string

get       any 'enable' or 'set' argument, or
          <all|tor_ssh_onion|tor_electrum_onion>

apply     no argument, applies all configuration settings to the system 
          [not yet implemented]
```

### Storage of configuration values
Settings are stored in individual files named in `/data/sysconfig/` as key/value pairs, named like the KEY in uppercase. 
For example, the file `BITCOIN_NETWORK` contains `BITCOIN_NETWORK=mainnet`. 
These files are always overwritten completely and can be sourced by any script so that the variable `BITCOIN_NETWORK` is available immediately.

```
$ ls -l /data/sysconfig/
  ... AUTOSETUP_SSD
  ... BITCOIN_NETWORK
  ... DASHBOARD_HDMI
  ... DASHBOARD_WEB
  ... WIFI

$ cat /data/sysconfig/BITCOIN_NETWORK 
BITCOIN_NETWORK=mainnet

$ source /data/sysconfig/BITCOIN_NETWORK 
$ echo $BITCOIN_NETWORK
mainnet
```

A backup/restore process simply copies all files within `/data/sysconfig/` to/from a different location. 
To apply a restored configuration, the `bbb-config.sh apply` command is executed.

### Managing configuration through the BitBox App
Ultimately, the configuration is managed by the user through the BitBox App, that talks to the Middleware which in turn calls the `bbb-config.sh` script with the necessary arguments. 
This has not been implemented yet.