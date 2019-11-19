---
layout: default
title: Build on device
parent: Tinkering
nav_order: 100
---
## Build on device

The BitBoxBase can also be built directly on a device. This is more suited for development or to create a "golden image" on different distributions.

### Supported environments

For example, it can run directly on a RockPro64 with a new Ayufan distribution.
It cannot be guaranteed that everything works, as for the moment, we only test on the RockPro64 and the following distributions:

* Armbian Ubuntu 18.04
* Ayufan Ubuntu 18.04

Other boards and Debian-based distributions could work as well.

### Run installation

* Setup your board and Linux distribution and log in using SSH with `root` or a sudo user.
* Clone the BitBoxBase GitHub repository and start the installation
    ```sh
    git clone https://github.com/digitalbitbox/bitbox-base.git
    cd bitbox-base/armbian
    sudo make build-ondevice
    ```
* When prompted with the configuration file, make desired changes and save/exit with `Ctrl-O`, `Enter`, `Ctrl-X`.
  See ["Operating System / Build Armbian"](../os/armbian-build.md#initial-configuration-on-build) for additional information.
* Reboot and connect with the BitBoxApp
