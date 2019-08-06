---
layout: default
title: Architecture
parent: Overview
nav_order: 110
---
## Architecture

The BitBox Base integrates seamlessly with the [BitBox App](https://shiftcrypto.ch/app/), which functions as control center for all node functionality, and supported hardware wallets.
The two components discover each other within a local network without manual configuration and can then reconnect after initial pairing using different connection methods.

See dedicated documentation sections on the left for additional details.

### User interface

The BitBox Base runs as a headless appliance with a minimal status display.
It is used and managed through the free and open-source BitBox App.
Having the user interface in a seperate application simplifies many things, allowing for automatic network discovery, a setup wizard and secure remote management.
This apporach also reduces the attack surface significantly, as no webserver needs to be exposed and port-forwarding can be avoided completely.

The BitBox App is hosted in a seperate GitHub repository:
<https://github.com/digitalbitbox/bitbox-wallet-app>

### Hardware

Building a solution platform that focuses on security and performance, the BitBox Base uses an ARM-based board with enough processing power to enable additional features in the future.

* [Pine64 ROCKPro64](https://www.pine64.org/rockpro64/) with fast 4GB memory and an internal 1TB SSD
* BitBox secure chip: adapted BitBox 02 that drives trusted screen and buttons

### Operating system

The operating system is custom-built as a minimal firmware, running in read-only mode and allowing atomic updates with fallback.

* [Armbian](https://www.armbian.com/): custom built Linux operating system, mounted as read-only with tmpfs overlayfs from eMMC storage
* [Mender.io](https://mender.io/): Over-the-air update management solution, enabling atomic full diskimage updates, using dual partitions for fallback

### Applications

The following key applications are used:

* [Bitcoin Core](https://bitcoincore.org/): full Bitcoin node, communicating directly with the peer-to-peer network, validating and broadcasting transactions
* [c-lightning](https://github.com/ElementsProject/lightning/blob/master/README.md): Lightning Network client specifically built for backend usage
* [electrs](https://github.com/romanz/electrs/blob/master/README.md): Electrum Server to provide blockchain data to software wallets

The following services are exposed:

* [NGINX](https://www.nginx.com/): reverse proxy to handle all incoming traffic
* [Base Middleware](https://github.com/digitalbitbox/bitbox-base/tree/master/middleware): custom middleware managing encrypted communication between BitBox Base and App

Additional noteworthy components on the BitBox Base:

* [Base Supervisor](https://github.com/digitalbitbox/bitbox-base/tree/master/tools/bbbsupervisor): custom daemon for operational monitoring and control, providing system health information and node configuration
* [Tor](https://www.torproject.org/): external network connections exclusively use the privacy-focused Tor network
* [Prometheus](https://prometheus.io/): monitoring of system and software components
* [Grafana](https://grafana.com/): visualization of system and network performance metrics

### Networking

Connectivity from the Bitcoin wallet application to the node backend is a challenge. We provide the following complementary options to allow for privacy and ease-of-use:

1. Local network: automatic detection using mDNS within the local network.
2. Tor network: private connectivity without any router configuration, needs Tor installed on client device.  
3. Shift Connect: zero-knowledge Tor/Web proxy operated by Shift for use with any client device

Overall, we strive to make using our BitBox products as simple as possible.
