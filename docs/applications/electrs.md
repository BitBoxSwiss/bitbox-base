---
layout: default
title: electrs
parent: Applications
nav_order: 120
---
## electrs

electrs is an Electrum Server implementation in Rust, with the project information and source code available on GitHub ([romanz/electrs](https://github.com/romanz/electrs)).
It uses Bitcoin Core as a data source and builds its own set of indexes to server Bitcoin transaction data to Electrum clients, like our BitBoxApp.

### Building from source or pre-compiled binary

The electrs project does not provide official binary releases.
Because compiling from source during the Armbian build fails due to timeouts, a [precompiled .deb package](https://github.com/digitalbitbox/bitbox-base-deps) is used.
This is a temporary solution and we'd like to address that and contribute to a solution over time.

### Configuration

The application configuration is specified in the local `/etc/electrs/electrs.conf` file. Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/customize-armbian-rockpro64.sh).

```
NETWORK=testnet
RPCCONNECT=127.0.0.1
RPCPORT=18332
DB_DIR=/mnt/ssd/electrs/db
DAEMON_DIR=/mnt/ssd/bitcoin/.bitcoin
MONITORING_ADDR=127.0.0.1:4224
VERBOSITY=vvvv
RUST_BACKTRACE=1
```

Contrary to other applications, this config file defines environment variables that are passed to electrs as commandline arguments.

Additional information about commandline arguments can be found by running `electrs --help`:

```
Electrum Rust Server 0.6.1

USAGE:
    electrs [FLAGS] [OPTIONS]

FLAGS:
    -h, --help              Prints help information
        --jsonrpc-import    Use JSONRPC instead of directly importing blk*.dat files. Useful for remote full node or low
                            memory system
        --timestamp         Prepend log lines with a timestamp
    -V, --version           Prints version information
    -v                      Increase logging verbosity

OPTIONS:
        --bulk-index-threads <bulk_index_threads>
            Number of threads used for bulk indexing (default: use the # of CPUs) [default: 0]

        --cookie <cookie>
            JSONRPC authentication cookie ('USER:PASSWORD', default: read from ~/.bitcoin/.cookie)

        --daemon-dir <daemon_dir>                    Data directory of Bitcoind (default: ~/.bitcoin/)
        --daemon-rpc-addr <daemon_rpc_addr>
            Bitcoin daemon JSONRPC 'addr:port' to connect (default: 127.0.0.1:8332 for mainnet, 127.0.0.1:18332 for
            testnet and 127.0.0.1:18443 for regtest)
        --db-dir <db_dir>                            Directory to store index database (default: ./db/)
        --electrum-rpc-addr <electrum_rpc_addr>
            Electrum server JSONRPC 'addr:port' to listen on (default: '127.0.0.1:50001' for mainnet, '127.0.0.1:60001'
            for testnet and '127.0.0.1:60401' for regtest)
        --index-batch-size <index_batch_size>
            Number of blocks to get in one JSONRPC request from bitcoind [default: 100]

        --monitoring-addr <monitoring_addr>
            Prometheus monitoring 'addr:port' to listen on (default: 127.0.0.1:4224 for mainnet, 127.0.0.1:14224 for
            testnet and 127.0.0.1:24224 for regtest)
        --network <network>                          Select Bitcoin network type ('mainnet', 'testnet' or 'regtest')
        --server-banner <server_banner>
            The banner to be shown in the Electrum console [default: Welcome to electrs (Electrum Rust Server)!]

        --tx-cache-size <tx_cache_size>
            Number of transactions to keep in for query LRU cache [default: 10000]

        --txid-limit <txid_limit>
            Number of transactions to lookup before returning an error, to prevent "too popular" addresses from causing
            the RPC server to get stuck (0 - disable the limit) [default: 100]
```

### Service management

The bitcoind service is managed by systemd. Relevant parameters are specified in the unit file `electrs.service` shown below.
Please check the most current initial configuration in [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/customize-armbian-rockpro64.sh).

```
[Unit]
Description=Electrs server daemon
Wants=bitcoind.service
After=bitcoind.service

[Service]
ExecStartPre=/bin/systemctl is-active bitcoind.service
ExecStart=/opt/shift/scripts/systemd-start-electrs.sh
RuntimeDirectory=electrs
User=electrs
Group=bitcoin
Type=simple
KillMode=process
Restart=always
TimeoutSec=120
RestartSec=30
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
* `ExecStart`: instead of starting electrs directly, it is called with a shell script to allow for additional commands (see next paragraph)
* `User`: runs as service user "electrs"
* `Group`: use group "bitcoin" to have access to the relevant bitcoind directories
* `Restart`: always restarted, unless manually stopped
* `PrivateTmp`: using a private tmp directory
* `ProtectSystem`: mount /usr, /boot/ and /etc read-only for the process
* `NoNewPrivileges`: disallow the process and all of its children to gain new privileges through execve()
* `PrivateDevices`: use a new /dev namespace only populated with API pseudo devices such as /dev/null, /dev/zero and /dev/random
* `MemoryDenyWriteExecute`: deny the creation of writable and executable memory mappings

## Starting electrs

The systemd unit executes electrs with the shell script [`systemd-electrs-startpre.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/systemd-electrs-startpre.sh). This allows the execution of additional commands for preparation and post-processing.

For additional details, please check the inline comments directly in the script.

## Data storage

The additional indexes are stored on the SSD due to their significant size. The storage location `/mnt/ssd/electrs/db` is specified in the `electrs.conf` file.
