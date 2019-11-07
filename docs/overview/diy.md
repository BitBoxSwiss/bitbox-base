---
layout: default
title: Do It Yourself!
parent: Overview
nav_order: 140
---
## Do It Yourself!

The BitBoxBase projects encourages you to build your own Bitcoin full node! It is still under heavy development and not ready for primetime, so this section will become more detailed over time.

### Hardware assembly

Although we use custom hardware to improve our commercial product and enable additional features, you can build your own full node using standard components:

* Pine64 components
  * [RockPRO64 4GB](https://store.pine64.org/?product=rockpro64-4gb-single-board-computer) single board computer
  * [eMMC 16 GB](https://store.pine64.org/?product=16gb-emmc) and [USB adapter](https://store.pine64.org/?product=usb-adapter-for-emmc-module)
  * [Power adapter EU](https://store.pine64.org/?product=rockpro64-12v-3a-eu-power-supply) or [US](https://store.pine64.org/?product=rockpro64-12v-3a-us-power-supply)
* Case option A (internal storage)
  * Custom case (not available yet)
  * [PCIe SSD adapter](https://store.pine64.org/?product=rockpro64-pci-e-x4-to-m-2ngff-nvme-ssd-interface-card) + 1 TB SSD
  * [Mid Profile Heatsink](https://store.pine64.org/?product=rockpro64-20mm-mid-profile-heatsink)
  * [CPU Fan](https://store.pine64.org/?product=fan-for-rockpro64-20mm-mid-profile-heatsink)
* Case option B (external storage)
  * [Pine64 Aluminium Casing](https://store.pine64.org/?product=rockpro64-premium-aluminum-casing)
  * external USB3 drive (SSD or HDD)

### Build Base image

The automated build process compiles the custom Armbian operating system, installs and configures all applications and prepares the image for Mender OTA updates (optional).

**Prerequisites**

Make sure you have the following prerequisites installed on your computer. At the moment, we test the whole process on Ubuntu only.

* Docker CE, version >= 18.06.3
  install manually according to [the official documentation](https://docs.docker.com/install/)

* [Git](https://git-scm.com/) and `qemu-user-static`
  ```
  sudo apt-get install git qemu-user-static
  ```

**Compile Armbian and custom applications**

The BitBoxBase runs a minimal Armbian operating system with additional custom applications written by Shift Cryptosecurity.
The main output is an Armbian image that contains the compiled custom applications and can be used to boot the BitBoxBase.

We assume that running Docker requires `sudo`, therefore `sudo make` is needed. If your Docker installation allows execution for regular users, `sudo` is not necessary.

* Building the BitBoxBase system image
  ```bash
  sudo make
  ```

* Optional: updating the BitBoxBase system image later with an adjusted build configuration
  ```bash
  sudo make update
  ```

See [Building the Armbian base image](/os/armbian-build.md) and related pages for more details.

**Create Mender.io update artefacts**

The Armbian disk image contains only one partition and cannot be updated remotely.
To integrate the BitBoxBase with the professional Mender.io update management solution, this image is postprocessed.
The result is a disk image with multiple partition that contain the mender configuration and allow over-the-air updates.

* Creating Mender.io disk image and update artefacts based on the Armbian system image
  ```bash
  sudo make mender-artefacts
  ```

### Assembly

(TODO)Stadicus

### Operations

(TODO)Stadicus
