---
layout: default
title: Bitcoin Core
parent: Applications
nav_order: 100
---
## Bitcoin Core

Bitcoin Core the most popular software implementation of Bitcoin, openly developed on GitHub (<https://github.com/bitcoin/bitcoin>) and publishing releases on <https://bitcoincore.org>. The BitBoxBase uses the latest stable binary release to communicate with the Bitcoin peer-to-peer network, for example to learn about current transactions and receive newly mined blocks. All data is regarded as untrusted until it is locally verified to match the consensus rules of Bitcoin Core.

### Using binary releases

While the `master` branch on GitHub reflects the current development process, a new stable release is tagged and tested every few months. This release is built from the source code in a reproducible way for various platforms by many contributors, verifying that they get the exact same binary executables by comparing the sha256 hash. These hash values for all images are provided in the `SHA256SUM.asc` file, which is signed by a Bitcoin Core maintainer, currently [laanwj](https://github.com/laanwj).

Instead of compiling Bitcoin Core from source when building the Armbian image, or poviding our own binaries, we believe that following the official binary releases is a good way to minimize trusting any single entity and keep the build process performant.

To ensure that the official binary is used, we store the verified signing key `01EA 5486 DE18 A882 D4C2  6845 90C8 019E 36C2 E964` independently from the Bitcoin Core release site and validate the downloaded binary against the signed hash values. If this verification does not succeed (a single bit difference would be enough) the build script aborts with an error.

### Configuration

The application configuration is specified in the local `/etc/bitcoin/bitcoin.conf` file. The `bitcoin.conf` file is generated from a template during the build process of the Armbian image. It can updated during regular operations. These are the initial settings:

```console
# network
mainnet=1
testnet=0

# server
server=1
listen=1
listenonion=1
daemon=0
txindex=0
prune=0
disablewallet=1
printtoconsole=1

# rpc
rpcconnect=127.0.0.1
rpcport=8332
rpcauth=xxx

# performance
dbcache=2000
maxconnections=40
maxuploadtarget=5000

# tor
proxy=127.0.0.1:9050
seednode=nkf5e6b7pl4jfd4a.onion
seednode=xqzfakpeuvrobvpj.onion
seednode=tsyvzsqwa2kkf6b2.onion

# validation
reindex-chainstate=0
```

Some notes about this specific configuration:

* **Network options**
  * `mainnet`/`testnet`: the build script defaults to building a mainnet node, but can be reconfigured by:
    * specifying the corresponding build parameter in [`build.conf`](../../armbian/base/build/build.conf)
    * or running the command `bbb-config.sh set bitcoin_network testnet` manually on the BitBoxBase (see [Operating System/Helper Scripts](../os/helper-scripts.md)).

* **Server options**
  * `server`: enables the RPC interface
  * `listen`: accept connections from other nodes
  * `listenonion`: create a Tor hidden service and accept incoming connections from other nodes on that address
  * `daemon`: to caputre all log output with the `printtoconsole` option, bitcoind must not run as a daemon
  * `txindex`: a full transaction index is not necessary, as we use electrs with its own indices to serve transaction information
  * `prune`: both c-lightning as well as electrs do not fully support pruned Bitcoin nodes at the moment
  * `disablewallet`: wallet functionality of bitcoind is not used and therefore disabled

* **RPC options**
  * `rpconnect`: specifying `127.0.0.1`, the bitcoind api listens only for local connections
  * `rpcport`: the bitcoind api always listens on port `8334`, both for mainnet and testnet
  * `rpcauth`: authentication to bitcoind api uses the `rpcauth` method, with clients using static `rpcuser` and `rpcpassword` values. Credentials are created using an [adapted version](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bitcoind-rpcauth.py) of the Bitcoin Core [rpcauth.py](https://github.com/bitcoin/bitcoin/tree/master/share/rpcauth) Python script

* **Performance options**
  * `dbcache`: initially, the database cache is optimized for the initial block download (IBD), allocating 2GB memory. This value is set to a lower value (default: 300 MB) after bitcoind is fully synced.
  * `maxconnections`: upper limit number of connections to other Bitcoin nodes
  * `maxuploadtarget`: upper limit for daily upload data volume

* **Tor**
  * `proxy`: set bitcoind to also use the Tor proxy for IPv4 and IPv6 connections
  * `seednode`: initial Tor nodes to bootstrap connections and discover additional nodes

* **Validation**
  * `reindex-chainstate`: if this option is set to `1`, e.g. by the script [`bbb-config.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-config.sh), bitcoind discards the current UTXO set on restart and revalidates the whole blockchain from Genesis.

### Service management

The bitcoind service is managed by systemd. Relevant parameters are specified in the unit file `bitcoind.service` shown below. Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](../../armbian/base/build/customize-armbian-rockpro64.sh).

```console
[Unit]
Description=Bitcoin daemon
After=multi-user.target redis.service
Requires=startup-checks.service

[Service]

# Service execution
###################

# run startpre script as root (with =+ )
ExecStartPre=+/opt/shift/scripts/systemd-bitcoind-startpre.sh

ExecStart=/usr/bin/bitcoind \
    -conf=/etc/bitcoin/bitcoin.conf \
    -pid=/run/bitcoind/bitcoind.pid
ExecStartPost=/opt/shift/scripts/systemd-bitcoind-startpost.sh

# Process management
####################
# bitcoind is run as 'simple', as otherwise log output is not caputred by journald
Type=simple
PIDFile=/run/bitcoind/bitcoind.pid
Restart=always
RestartSec=10
TimeoutSec=300


# Directory creation and permissions
####################################

# Run as bitcoin:bitcoin
User=bitcoin
Group=bitcoin

# /run/bitcoind
RuntimeDirectory=bitcoind
RuntimeDirectoryMode=0710

# /etc/bitcoin
ConfigurationDirectory=bitcoin
ConfigurationDirectoryMode=0755

# /var/lib/bitcoind
StateDirectory=bitcoind
StateDirectoryMode=0710


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

* `After`: started after regular Linux start and after Redis is available.
* `Requires`: the `systemd-startup-checks.sh` must run successfully.
* `ExecStartPre`: a preparation script is run as root (see `=+` option) to check Redis availability, valid RPCAUTH credentials and access to the SSD
* `ExecStart`: starts `bitcoind`
* `User` / `Group`: runs as service user "bitcoin"
* `Restart`: always restarted, unless manually stopped
* `PrivateTmp`: using a private tmp directory
* `ProtectSystem`: mount /usr, /boot/ and /etc read-only for the process
* `NoNewPrivileges`: disallow the process and all of its children to gain new privileges through execve()
* `PrivateDevices`: use a new /dev namespace only populated with API pseudo devices such as /dev/null, /dev/zero and /dev/random
* `MemoryDenyWriteExecute`: deny the creation of writable and executable memory mappings
* `WantedBy`: custom applications are executed after Linux system boot (target `multi-user.target`) for the custom target `bitboxbase.target`


## Data storage

As bitcoind is run as user "bitcoin", it uses the user's home directory (set to the SSD as `/mnt/ssd/bitcoin/`) as the standard data directory and creates the directory `.bitcoin` at this location.
