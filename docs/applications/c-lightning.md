---
layout: default
title: c-lightning
parent: Main Applications
nav_order: 620
---
## c-lightning

c-lightning is a Lightning Network client / server implementation in C by Blockstream, with the project information and source code available on GitHub ([ElementsProject/lightning](https://github.com/ElementsProject/lightning)).

### Building from source or pre-compiled binary

The c-lightning project does provide official binary releases, but not for the Arm64 platform.
For production releases, we therefore compile it during the Armbian build directly from source by default.
As this takes a while, we also provide a [precompiled .deb package](https://github.com/digitalbitbox/bitbox-base-deps) that can be used by the Armbian build process, but this is recommended for development images only.
Usage of the precompiled image can be enabled in [`build.conf`](../../armbian/base/build/build.conf).

### Configuration

The application configuration is specified in the local `/etc/lightningd/lightningd.conf` file. Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](../../armbian/base/build/customize-armbian-rockpro64.sh).

```
bitcoin-cli=/usr/bin/bitcoin-cli
bitcoin-rpcconnect=127.0.0.1
bitcoin-rpcport=18332
network=testnet
lightning-dir=/mnt/ssd/bitcoin/.lightning-testnet
bind-addr=127.0.0.1:9735
proxy=127.0.0.1:9050
log-level=debug
daemon
plugin=/opt/shift/scripts/prometheus-lightningd.py
```

Some notes about this specific configuration:

* `bitcoin-rpcconnect` / `bitcoin-rpcport`: ip address and port to connect to the bitcoind RPC
* `lightning-dir`: to allow switching between independent mainnet / testnet c-lightning instances, a seperate directory for testnet is used
* `bind`: only listen to local connections (e.g. through Tor or the middleware)
* `proxy`: connect to the Tor proxy for all external communication
* `plugin`: start the `prometheus-lightningd.py` plugin to provide metrics to Prometheus monitoring

Additional information can be found in the reference [lightningd.config](https://github.com/ElementsProject/lightning/blob/master/doc/lightningd-config.5.txt).

### Service management

The bitcoind service is managed by systemd. Relevant parameters are specified in the unit file `lightningd.service` shown below.
Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](../../armbian/base/build/customize-armbian-rockpro64.sh).

```
[Unit]
Description=c-lightning daemon
Wants=bitcoind.service
After=bitcoind.service

[Service]
ExecStartPre=/bin/systemctl is-active bitcoind.service
ExecStart=/opt/shift/scripts/systemd-start-lightningd.sh
RuntimeDirectory=lightningd
User=bitcoin
Group=bitcoin
Type=forking
Restart=always
RestartSec=10
TimeoutSec=240
PrivateTmp=true
ProtectSystem=full
NoNewPrivileges=true
PrivateDevices=true
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
```

Some notes about this specific configuration:

* `After`: started after bitcoind
* `ExecStartPre`: checks if systemd is running and fails if that's not the case
* `ExecStart`: instead of starting lightningd directly, it is called with a shell script to allow for additional commands (see next paragraph)
* `User`: runs as service user "bitcoin"
* `Restart`: always restarted, unless manually stopped
* `PrivateTmp`: using a private tmp directory
* `ProtectSystem`: mount /usr, /boot/ and /etc read-only for the process
* `NoNewPrivileges`: disallow the process and all of its children to gain new privileges through execve()
* `PrivateDevices`: use a new /dev namespace only populated with API pseudo devices such as /dev/null, /dev/zero and /dev/random
* `MemoryDenyWriteExecute`: deny the creation of writable and executable memory mappings

## Starting c-lightning

The systemd unit executes c-lightning with the shell script [`systemd-start-lightningd.sh`](../../armbian/base/scripts/systemd-start-lightningd.sh). This allows the execution of additional commands for preparation and post-processing.

For additional details, please check the inline comments directly in the script.

## Data storage

In the configuration file, either `/mnt/ssd/bitcoin/.lightning` or `/mnt/ssd/bitcoin/.lightning-testnet` is specified, depending on the Bitcoin network.
