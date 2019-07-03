---
layout: default
title: Getting started
nav_order: 150
has_children: true
permalink: /start
---
## BitBox Base: Getting started

This document provides a high-level overview of all steps involved to build a runnining BitBox Base yourself.

### Hardware

Most hardware components are readily available and you'll be able to assemble a working Base with these.
We use custom parts to improve the overall product and enable additional features.

Please refer to the [Hardware](hardware.md) section for a list of necessary parts.

### Operating System

Building all necessary software is mostly automated.

## Prerequisites

Make sure you have the following prerequisites installed on your computer. At the moment, we test the whole process on Ubuntu only.

* Docker CE, version >= 18.06.3  
  install manually according to [the official documentation](https://docs.docker.com/install/)
  
* [Git](https://git-scm.com/) and `qemu-user-static`  

  ```bash
  sudo apt-get install git qemu-user-static
  ```

#### Build Armbian operating system image

The BitBox Base runs a minimal Armbian operating system with additional custom applications written by Shift Cryptosecurity.
The main output is an Armbian image that contains the compiled custom applications and can be used to boot the BitBox Base.

We assume that running Docker requires `sudo`, therefore `sudo make` is needed. If your Docker installation allows execution for regular users, `sudo` is not necessary.

* Building the BitBox Base system image  

  ```bash
  sudo make
  ```

* Optional: updating the BitBox Base system image later with an adjusted build configuration  

  ```bash
  sudo make update
  ```

See [Building the Armbian base image](/os/armbian-build.md) and related pages for more details.


#### Create Mender.io update artefacts

The Armbian disk image contains only one partition and cannot be updated remotely.
To integrate the BitBox Base with the professional Mender.io update management solution, this image is postprocessed.
The result is a disk image with multiple partition that contain the mender configuration and allow over-the-air updates.

* Creating Mender.io disk image and update artefacts based on the Armbian system image  

  ```bash
  sudo make mender-artefacts
  ```

## Assembly

(TODO)Stadicus

## Operations

(TODO)Stadicus
