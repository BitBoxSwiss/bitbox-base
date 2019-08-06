---
layout: default
title: Bitcoin Core
parent: Applications
nav_order: 100
---
## Bitcoin Core

Bitcoin Core the most popular software implementation of Bitcoin, openly developed on GitHub (<https://github.com/bitcoin/bitcoin>) and publishing releases on <https://bitcoincore.org>. The BitBox Base uses the latest stable binary release to communicate with the Bitcoin peer-to-peer network, for example to learn about current transactions and receive newly mined blocks. All data is regarded as untrusted until it is locally verified to match the consensus rules of Bitcoin Core.

### Using binary releases

While the `master` branch on GitHub reflects the current development process, a new stable release is tagged and tested every few months. This release is built from the source code in a reproducible way for various platforms by many contributors, verifying that they get the exact same binary executables by comparing the sha256 hash. These hash values for all images are provided in the `SHA256SUM.asc` file, which is signed by a Bitcoin Core maintainer, currently [laanwj](https://github.com/laanwj).

Instead of compiling Bitcoin Core from source when building the Armbian image, or poviding our own binaries, we believe that following the official binary releases is a good way to minimize trusting any single entity and keep the build process performant.

To ensure that the official binary is used, we store the verified signing key `01EA 5486 DE18 A882 D4C2  6845 90C8 019E 36C2 E964` independently from the Bitcoin Core release site and validate the downloaded binary against the signed hash values. If this verification does not succeed (a single bit difference would be enough) the build script aborts with an error.

### Configuration

The application configuration is specified in the local `/etc/bitcoin/bitcoin.conf` file. The `bitcoin.conf` file is written into the Armbian image by the  [`customize-armbian-rockpro64.sh`](../../armbian/base/build/customize-armbian-rockpro64.sh) script.

```
# application
testnet=1
server=1
listen=1
listenonion=1
daemon=1
txindex=0
prune=0
disablewallet=1
pid=/run/bitcoind/bitcoind.pid
rpccookiefile=/mnt/ssd/bitcoin/.bitcoin/.cookie
sysparms=1
rpcconnect=127.0.0.1

# performance
dbcache=2000
maxmempool=50
maxconnections=40
maxuploadtarget=5000

# tor
proxy=127.0.0.1:9050
seednode=nkf5e6b7pl4jfd4a.onion
seednode=xqzfakpeuvrobvpj.onion
seednode=tsyvzsqwa2kkf6b2.onion
```

Some notes about this specific configuration:

* **Application options**
  * `testnet`: the build script defaults to building a testnet node, but can be configured to run on mainnet (commenting out this line):
    * by specifying the corresponding build parameter in [`build.conf`](../../armbian/base/build/build.conf)
    * or by running the command `bbb-config.sh set bitcoin_network mainnet` manually on the BitBox Base (see [OS/Helper Scripts](../os/helper-scripts.md)).
  * `server`: enables the RPC interface
  * `listen`: accept connections from other nodes
  * `listenonion`: create a Tor hidden service and accept incoming connections from other nodes on that address
  * `daemon`: run bitcoind as a server in the background
  * `txindex`: a full transaction index is not necessary, as we use electrs - which has its own indices - to serve transaction information.
  * `prune`: both c-lightning as well as electrs do not support pruned Bitcoin nodes at the moment.
  * `disablewallet`: wallet functionality of bitcoind is not used and therefore disabled
  * `pid`: this file contains the process id of bitcoind
  * `rpccookiefile`: other applications using the bitcoind RPC authenticate themselves by reading the access credentials from the `.cookie` file. To avoid using different storage locations when running Bitcoin mainnet or testnet, the location is explicitly specified.
  * `sysparms`: other system users that are a member of the `bitcoin` group need to have read-only access to the `.cookie` file. Using this option, bitcoind creates new files with system default permissions instead of the umask 0777.
  * `rpcconnect`: as the bitcoind rpc port is sensitive, it listens for local connections only.
* **Performance options**
  * `dbcache`: initially, the database cache is optimized for the initial block download (IBD), allocating 2GB memory. This value is set to a lower value (default: 300 MB) after bitcoind is fully synced.
  * `maxmempool`: upper limit in MB for mempool
  * `maxconnections`: upper limit number of connections to other Bitcoin nodes
  * `maxuploadtarget`: upper limit for daily upload data volume
* **Tor**
  * `proxy`: set bitcoind to also use the Tor proxy for IPv4 and IPv6 connections
  * `seednode`: initial Tor nodes to bootstrap connections and discover additional nodes

### Service management

The bitcoind service is managed by systemd. Relevant parameters are specified in the unit file `bitcoind.service` shown below. Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](../../armbian/base/build/customize-armbian-rockpro64.sh).

```
[Unit]
Description=Bitcoin daemon
After=network-online.target startup-checks.service tor.service
Requires=startup-checks.service

[Service]
ExecStart=/opt/shift/scripts/systemd-start-bitcoind.sh
RuntimeDirectory=bitcoind
User=bitcoin
Group=bitcoin
Type=forking
PIDFile=/run/bitcoind/bitcoind.pid
Restart=always
RestartSec=60
TimeoutSec=300
PrivateTmp=true
ProtectSystem=full
NoNewPrivileges=true
PrivateDevices=true
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
```

Some notes about this specific configuration:

* `After`: started after network is established, Tor is enabled and startup-checks have passed.
* `ExecStart`: instead of starting bitcoind directly, it is called with a shell script to allow for additional commands (see next paragraph)
* `User`: runs as service user "bitcoin"
* `Restart`: always restarted, unless manually stopped
* `PrivateTmp`: using a private tmp directory
* `ProtectSystem`: mount /usr, /boot/ and /etc read-only for the process
* `NoNewPrivileges`: disallow the process and all of its children to gain new privileges through execve()
* `PrivateDevices`: use a new /dev namespace only populated with API pseudo devices such as /dev/null, /dev/zero and /dev/random
* `MemoryDenyWriteExecute`: deny the creation of writable and executable memory mappings

## Starting Bitcoin Core

The systemd unit executes bitcoind with the shell script [`systemd-start-bitcoind.sh`](../../armbian/base/scripts/systemd-start-bitcoind.sh). This allows the execution of additional commands for preparation and post-processing.

For additional details, please check the inline comments directly in the script.

## Data storage

As bitcoind is run as user "bitcoin", it uses the user's home directory (set to the SSD as `/mnt/ssd/bitcoin/`) as the standard data directory and creates the directory `.bitcoin` at this location.
