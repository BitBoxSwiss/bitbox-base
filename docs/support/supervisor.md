---
layout: default
title: Base Supervisor
parent: Supporting Applications
nav_order: 710
---
## Base Supervisor: Custom System Management

Every running system needs to be managed.
On a services level, this is the job of `systemd` which starts application services in the right order, keeps track of logs and restarts services if they crash.
On top of this very mechanic management, a managment service is needed that actually understands the many inticacies of the various applications.
The Base Supervisor `bbbsupervisor` is custom-built to monitor application logs and other system metrics, watches for very specific application messages, knows how to interpret them and can take the required action.

### Scope

The Base Supervisor combines many small monitoring tasks. Contrary to the Middleware, its task is not about relaying application communication but to keep the running system in an operational state without user interaction.

* **System**
  * temperature control: monitor bbbfancontrol and throttle CPU if needed
  * disk space: monitor free space on rootfs and ssd, perform cleanup of temp & logs
  * memory: detect memory issues or "zram decompression failed", perform reboot
  * swap: detect issues with swap file

* **Middleware**
  * monitor service availability

* **Bitcoin Core**
  * monitor service availability
  * perform backup tasks using hardlinks
  * switch between IBD and normal operation mode (e.g. adjust `dbcache`)

* **c-lightning**
  * monitor service availability
  * perform backup tasks (once possible)

* **electrs**
  * monitor service availability
  * track initial sync and full database compaction, restart service if needed

* **NGINX, Grafana, ...**
  * monitor service availability

This list is non-exhaustive and likely to grow.

🏗️ [WIP](https://github.com/shiftdevices/bitbox-base-internal/issues/142): the Base Supervisor is under heavy development and not functional yet.

### Installation

The [application](../../tools/bbbsupervisor/bbbsupervisor.go) is written in Go, compiled within Docker when using the top `make` command and the resulting binary is copied to the Armbian image during build.

### Service management

The Base Supervisor is started and managed using a simple [systemd unit file](../../tools/bbbsupervisor/bbbsupervisor.service):

```
[Unit]
Description=BitBox Base Supervisor
After=local-fs.target

[Service]
Type=simple
ExecStart=/usr/local/sbin/bbbsupervisor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```
