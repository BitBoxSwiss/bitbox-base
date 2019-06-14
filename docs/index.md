---
layout: default
title: Home
nav_order: 100
---
# BitBox Base

The BitBox Base is an ongoing project of [Shift Cryptosecurity](https://shiftcrypto.ch/) that aims to build a personal Bitcoin full node appliance. The whole softare stack is free open-source. This documentation is aimed at project members, contributors and intersted people that want to reuse our work on their own devices.

## Table of Content

1. [About](about.md)
1. Hardware
   1. [Specifications Overview](hw/spec-overview.md)
   1. [Platform Choice](hw/platform-choice.md)
   1. [CAD Concept Schematics](hw/cad-concept-schematics.md)
1. Creating the Base image
   1. Build system
   1. Build configuration
   1. Dependencies
1. Operating System
   1. [Build Armbian image](os/armbian-build.md)
   1. [Build details](os/build-details.md)
   1. [Configuration](os/configuration.md)
   1. [Security considerations](os/security.md)
   1. [Helper scripts](os/helper-scripts.md)
1. Base Middleware
   1. API
   1. Noise encryption
   1. HSM integration
1. Main applications
   1. [Bitcoin Core](applications/bitcoin-core.md)
   1. [c-lightning](applications/c-lightning.md)
   1. [Electrs](applications/electrs.md)
1. Supporting applications
   1. Base Supervisor
   1. [Tor](support/tor.md)
   1. [NGINX](support/nginx.md)
   1. Custom tools
   1. Prometheus
   1. Grafana
1. Firmware upgrades
   1. [Overall concept](upgrade/concept.md)
   1. Device implementation
   1. Attestation
   1. Custom firmware
1. [Contributing](contributing.md)

## Contributor workflow

We are building the software stack of the BitBox Base fully open source and with its application outside of our own hardware device in mind. Contributions are very welcome. Please read the [Contributing](contributing.md) section before submitting changes to the repository.