---
layout: default
title: Home
nav_order: 100
---
# BitBox Base

## Personal Bitcoin sovereignty node

The BitBox Base is an ongoing project of [Shift Cryptosecurity](https://shiftcrypto.ch/) that aims to build a personal Bitcoin full node appliance.
The whole software stack is free open-source.
This documentation is aimed at project members, contributors and intersted people that want to build or customize their own node.

[View on GitHub](https://github.com/digitalbitbox/bitbox-base){: .btn } [Follow us on Twitter](https://twitter.com/ShiftCryptoHQ){: .btn }

## Table of Content

1. [Overview](overview/)
   1. [Architecture](overview/architecture.md)
   2. [Do It Yourself!](overview/diy.md)
2. [Hardware](hardware/)
   1. [Components](hardware/components.md)
   2. [UART communication](hardware/uart-communication.md)
3. [Operating System](os/)
   1. [Configuration](os/configuration.md)
   2. [Build Armbian image](os/armbian-build.md)
   3. [Security considerations](os/security.md)
   4. [Firmware upgrades](os/upgrade.md)
   5. [Common issues](os/os-faq.md)
4. [Custom applications](customapps/)
   1. [Middleware](customapps/bbbmiddleware.md)
   2. [Supervisor](customapps/bbbsupervisor.md)
   3. [Fan control](customapps/bbbfancontrol.md)
   4. [Helper scripts](customapps/helper-scripts.md)
   5. [Compiling binaries](customapps/go-build.md)
5. [Applications](applications/)
   1. [Bitcoin Core](applications/bitcoin-core.md)
   2. [c-lightning](applications/c-lightning.md)
   3. [Electrs](applications/electrs.md)
   4. [Tor](applications/tor.md)
   5. [NGINX](applications/nginx.md)
   6. [Prometheus](applications/prometheus.md)
   7. [Grafana](applications/grafana.md)
6. [Networking](networking/)
7. [Contributing](contributing/)

## Contributor workflow

We are building the software stack of the BitBox Base fully open source and with its application outside of our own hardware device in mind. Contributions are very welcome. Please read the [Contributing](contributing.md) section before submitting changes to the repository.
