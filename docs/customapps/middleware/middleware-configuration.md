---
layout: default
title: Configuration
parent: Middleware
grand_parent: Custom applications
nav_order: 200
---
## Middleware: Configuration


### Command line interface parameters

Running `bbbmiddleware -h` prints the configuration options for the Middleware on the command line.

Note: *The argument list below was last updated in December 2019.*

```console
-bbbcmdscript string
    Path to the bbb-cmd.sh script that allows executing system commands (default "/opt/shift/scripts/bbb-cmd.sh")
-bbbconfigscript string
    Path to the bbb-config.sh script that allows setting system configuration (default "/opt/shift/scripts/bbb-config.sh")
-bbbsystemctlscript string
    Path to the bbb-systemctl.sh script that allows starting and stopping services on the Base (default "/opt/shift/scripts/bbb-systemctl.sh")
-datadir string
    Directory where the Middleware persistent data, like for example the noise encryption keys, is stored (default ".base")
-electrsport string
    Electrs RPC port (default "51002")
-hsmfirmwarefile string
    Location of the signed HSM firmware binary (default "/opt/shift/hsm/firmware-bitboxbase.signed.bin")
-hsmserialport string
    Serial port used to communicate with the HSM (default "/dev/ttyS0")
-middlewareport string
    Port the Middleware listens on (default "8845")
-network string
    Indicate wether Bitcoin is running on mainnet or testnet (default "testnet")
-notificationNamedPipePath string
    Path where the Middleware creates a named pipe to receive notifications from other processes on the BitBoxBase (default "/tmp/middleware-notification.pipe")
-prometheusurl string
    URL of the Prometheus server (default "http://localhost:9090")
-redismock
    Flag to use the Redis mock for development instead of connecting to a redis server
-redisport string
    Port of the Redis server (default "6379")
-updatehsmfirmware
    Set to true to force HSM firmware update
-updateinfourl string
    URL to query information about Base image updates from (default "https://shiftcrypto.ch/updates/base.json")
```

### Changing the Middleware systemd file

(TODO)Stadicus

- config in systemd unit (via env variables / config template in /etc/bbbmiddleware)
