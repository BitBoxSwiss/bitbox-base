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

1. [Overview](overview.md)
   1. [Architecture](overview/architecture.md)
   2. [Do It Yourself!](overview/diy.md)
2. [Hardware](hw.md)
   1. [Specifications Overview](hw/spec-overview.md)
   2. [Platform Choice](hw/platform-choice.md)
   3. [CAD Concept Schematics](hw/cad-concept-schematics.md)
   4. [UART communication](hw/uart-communication.md)
3. [Operating System](os.md)
   3. [Configuration](os/configuration.md)
   1. [Build Armbian image](os/armbian-build.md)
   2. [Security considerations](os/security.md)
   3. [Common issues](os/os-faq.md)
4. [Custom applications](customapps.md)
   1. [Middleware](customapps/bbbmiddleware.md)
   2. [Supervisor](customapps/bbbsupervisor.md)
   3. [Fan control](customapps/bbbfancontrol.md)
   4. [Helper scripts](customapps/helper-scripts.md)
   5. [Compiling binaries](customapps/go-build.md)
5. [Applications](applications.md)
   1. [Bitcoin Core](applications/bitcoin-core.md)
   2. [c-lightning](applications/c-lightning.md)
   3. [Electrs](applications/electrs.md)
   4. [Tor](applications/tor.md)
   5. [NGINX](applications/nginx.md)
   6. [Prometheus](applications/prometheus.md)
   7. [Grafana](applications/grafana.md)
6. [Firmware upgrades](upgrade.md)
   1. [Overall concept](upgrade/concept.md)
   2. Device implementation
   3. Attestation
   4. Custom firmware
7. [Contributing](contributing.md)

## Contributor workflow

We are building the software stack of the BitBox Base fully open source and with its application outside of our own hardware device in mind. Contributions are very welcome. Please read the [Contributing](contributing.md) section before submitting changes to the repository.
