---
layout: default
title: Redis
parent: Applications
nav_order: 145
---
## Redis: key/value data store

The Redis key/value datastore is used to store all configuration data. It can be queried from all software components, be it from the command line, Bash, Python or Go with minimal overhead. See the [Operating System / Configuration](../os/configuration.md) section for additional details.

### Installation

Redis is installed using the standard Armbian package. The default systemd unit is disabled in favor of a custom one optimized for this project.

```bash
## install Redis
apt install -y --no-install-recommends redis
mkdir -p /data/redis/
chown -R redis:redis /data/redis/

### disable standard systemd unit
systemctl disable redis-server.service
systemctl mask redis-server.service
```

### Configuration

The default Redis configuration is stored in `/etc/redis/redis.conf` and remains unchanged, except for an `include` command at the end for [`/etc/redis/redis-local.conf`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/rootfs/etc/redis/redis-local.conf) which overwrites some settings.

```
# main config (unchanged) is located at /etc/redis/redis.conf
# accept connections only from clients running into the same computer
bind 127.0.0.1
port 6379

# run as a systemd unit
supervised systemd
daemonize yes

# loglevel: debug, verbose, notice or warning
loglevel notice

# database configuration
databases 1
dbfilename bitboxbase.rdb
dir /data/redis/

# store to disk every 60s
save 60 1

# various
always-show-logo no
```

### Initial data provisioning

The Redis datastore is already used within the Armbian build process and populated with the factory configuration settings.
All values are imported from [`armbian/base/config/redis/factorysettings.txt`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/config/redis/factorysettings.txt) using the [Redis mass insertion protocol](https://redis.io/topics/mass-insert) with the helper script [`armbian/base/scripts/redis-pipe.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/redis-pipe.sh).

```
SET base:version xxx
SET base:overlayroot:enabled 1
SET base:rootpasslogin:enabled 0
SET base:dashboard:web:enabled 1
SET base:dashboard:hdmi:enabled 0
SET base:autosetupssd:enabled 1
SET base:wifi:enabled 0
SET base:wifi:ssid none
SET base:wifi:password none
SET base:hostname bitbox-base

SET tor:base:enabled 1
SET tor:ssh:enabled 0
SET tor:electrs:enabled 1
SET tor:bbbmiddleware:enabled 1
                        
SET bitcoind:ibd 1
SET bitcoind:ibd-clearnet 0
SET bitcoind:network mainnet
SET bitcoind:testnet 0
...
...
```

### Service management

The Redis service is managed by systemd. Relevant parameters are specified in the unit file [`redis.service`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/rootfs/etc/systemd/system/redis.service) shown below.

```bash
[Unit]
Description=Redis In-Memory Data Store
After=multi-user.target
# original config in /lib/systemd/system/redis-server.service

[Service]

# Service execution
###################

ExecStart=/usr/bin/redis-server /etc/redis/redis.conf
ExecStop=/bin/kill -s TERM $MAINPID

# Process management
####################

Type=forking
Restart=always
Restart=always
RestartSec=10
TimeoutSec=30

# Directory creation and permissions
####################################

User=redis
Group=redis

[Install]
WantedBy=bitboxbase.target
```

## Data storage

Redis is an in-memory datastore. The data is dumped frequently (every 60 seconds on changes) and on demand to `/data/redis/bitboxbase.rdb`, from where it can be backed up or restored.
