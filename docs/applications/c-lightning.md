---
layout: default
title: c-lightning
parent: Applications
nav_order: 110
---
## c-lightning

c-lightning is a Lightning Network client / server implementation in C by Blockstream, with the project information and source code available on GitHub ([ElementsProject/lightning](https://github.com/ElementsProject/lightning)).

### Building from source or pre-compiled binary

The c-lightning project does provide official binary releases, but not for the Arm64 platform.
For production releases, we therefore compile it during the Armbian build directly from source by default.
As this takes a while, we also provide a [precompiled .deb package](https://github.com/digitalbitbox/bitbox-base-deps) that can be used by the Armbian build process, but this is recommended for development images only.
Usage of the precompiled image can be enabled in [`build.conf`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build.conf).

### Configuration

The application configuration is specified in the local `/etc/lightningd/lightningd.conf` file. Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/customize-armbian-rockpro64.sh).

```
mainnet
bitcoin-cli=/usr/bin/bitcoin-cli
bitcoin-rpcuser=base
bitcoin-rpcpassword=xxxxxxxxxxxxxxx
bitcoin-rpcconnect=127.0.0.1
bitcoin-rpcport=8332
lightning-dir=/mnt/ssd/bitcoin/.lightning
bind-addr=127.0.0.1:9735
proxy=127.0.0.1:9050
log-level=debug
plugin=/opt/shift/scripts/prometheus-lightningd.py
```

Some notes about this specific configuration:

* `bitcoin-rpcuser` / `bitcoin-rpcpassword`: bitcoind RPC credentials, as generated randomly by [`bitcoind-rpcauth.py`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bitcoind-rpcauth.py)
* `bitcoin-rpcconnect` / `bitcoin-rpcport`: ip address and port to connect to the bitcoind RPC
* `lightning-dir`: to allow switching between independent mainnet / testnet c-lightning instances, a seperate directory for testnet is used
* `bind`: only listen to local connections (e.g. through Tor or the middleware)
* `proxy`: connect to the Tor proxy for all external communication
* `plugin`: start the `prometheus-lightningd.py` plugin to provide metrics to Prometheus monitoring

Additional information can be found in the reference [lightningd.config](https://github.com/ElementsProject/lightning/blob/master/doc/lightningd-config.5.txt).

### Service management

The bitcoind service is managed by systemd. Relevant parameters are specified in the unit file `lightningd.service` shown below.
Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/customize-armbian-rockpro64.sh).

```bash
[Unit]
Description=c-lightning daemon
After=multi-user.target bitcoind.service

[Service]

# Service execution
###################

ExecStartPre=/opt/shift/scripts/systemd-lightningd-startpre.sh
ExecStart=/usr/local/bin/lightningd \
    --conf=/etc/lightningd/lightningd.conf
ExecStartPost=/opt/shift/scripts/systemd-lightningd-startpost.sh

# Process management
####################

Type=simple
Restart=always
RestartSec=30
TimeoutSec=240

# Directory creation and permissions
####################################

# Run as bitcoin:bitcoin
User=bitcoin
Group=bitcoin

# /run/lightningd
RuntimeDirectory=lightningd
RuntimeDirectoryMode=0710


# Hardening measures
####################

# Provide a private /tmp and /var/tmp.
PrivateTmp=true

# Mount /usr, /boot/ and /etc read-only for the process.
ProtectSystem=full

# Deny access to /home, /root and /run/user
ProtectHome=true

# Disallow the process and all of its children to gain
# new privileges through execve().
NoNewPrivileges=true

# Use a new /dev namespace only populated with API pseudo devices
# such as /dev/null, /dev/zero and /dev/random.
PrivateDevices=true

# Deny the creation of writable and executable memory mappings.
MemoryDenyWriteExecute=true


[Install]
WantedBy=bitboxbase.target
```

Some notes about this specific configuration:

* `After`: started after bitcoind
* `ExecStartPre`: runs [`systemd-lightningd-startpre.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/systemd-lightningd-startpre.sh) to check dependencies
* `ExecStartPost`: runs [`systemd-lightningd-startpost.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/systemd-lightningd-startpost.sh) to set permissions after starting
* `User`: runs as service user "bitcoin"
* `Restart`: always restarted, unless manually stopped
* `PrivateTmp`: using a private tmp directory
* `ProtectSystem`: mount /usr, /boot/ and /etc read-only for the process
* `NoNewPrivileges`: disallow the process and all of its children to gain new privileges through execve()
* `PrivateDevices`: use a new /dev namespace only populated with API pseudo devices such as /dev/null, /dev/zero and /dev/random
* `MemoryDenyWriteExecute`: deny the creation of writable and executable memory mappings

## Data storage

In the configuration file, either `/mnt/ssd/bitcoin/.lightning` or `/mnt/ssd/bitcoin/.lightning-testnet` is specified, depending on the Bitcoin network.
