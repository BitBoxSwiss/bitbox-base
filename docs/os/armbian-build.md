---
layout: default
title: Build Armbian image
parent: Operating System
nav_order: 310
---
## Building the Armbian base image

The goal of building the Armbian operating system from source is to create a highly configurable base image, with granular control over kernel features, that can be written on an eMMC or SD card and boots directly into an operational state.

The process to build the Armbian image for the RockPro64 board is mostly automated and follows [Armbian best practices](https://docs.armbian.com/Developer-Guide_Build-Preparation). It utilizes a VirtualBox machine with Ubuntu 18.04 LTS as a build environment, configured and run automatically using Vagrant. The following build instructions have been tested both on Debian-based Linux systems and within Windows PowerShell.

### Requirements

* Regular computer with x86/x64 architecture, 4GB+ memory, 4+ cores
* minimum of 25 GB free space on a fast drive, preferrably SSD
* any operating system that can run Virtualbox and Vagrant is supported

### Prepare build environment

To set up and run the build environment in the virtual machine, the following software needs to be installed:
* Git ([download](https://git-scm.com/))
* Virtualbox ([download](https://www.virtualbox.org/)), version > 5.1.12
* Vagrant ([download](https://www.vagrantup.com/)), version >= 2.2

All sources and build scripts are contained in this repository, which needs to be cloned locally. The following commands are executed in the command line, either the Linux terminal or Windows PowerShell.  

In Linux you can directly run `make`, while in Windows PowerShell you need to run the build script directly with `sh`. In the following instructions, Windows users just replace `make` with `sh .\build.sh`.

* Clone the BitBox Base repository to a local directory.
  ```
  $ git clone https://github.com/digitalbitbox/bitbox-base.git
  $ cd bitbox-base/armbian
  ```

* The virtual machine will be configured running on 4 cpu cores and 8GB memory. You can adjust these settings in the `build.sh` script.

### Configure build options

The build itself and the initial configuration of the Base image (e.g. hostname, Bitcoin network or Wifi credentials) can be configured within the configuration file [`armbian/base/build/build.conf`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/build.conf). 

To preserve a local configuration, you can copy the file to `build-local.conf`. 
This file is excluded from Git source control and overwrites options from `build.conf`.

### Include SSH keys

It is recommended to use SSH keys to access the Base image. 
You can include your own keys in the file [authorized_keys](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/authorized_keys). 
Please refer to [this article](https://confluence.atlassian.com/bitbucketserver/creating-ssh-keys-776639788.html) on how to create your own set of new keys.

### Compile Armbian from source

Now the operating system image can be built. The whole BitBox Base configuration is contained in [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/customize-armbian-rockpro64.sh) and executed in a `chroot` environment at the end of the build process.

* Start the initial build process.
  ```
  $ make
  ```

* The resulting image is available in `armbian-build/output/images/` and can be written to eMMC or SD card using a program like [Etcher](https://www.balena.io/etcher/). On the Linux command line you can use `dd`: once the target medium is connected to your computer, get the device name (e.g. `/dev/sdb`). Check it carefully, all data on this device will be lost!  
  ```
  $ lsblk
  $ sudo dd if=armbian-build/output/images/Armbian_5.77_Rockpro64_Debian_stretch_default_4.4.176.img of=/dev/sdb bs=64K conv=sync status=progress
  $ sync
  ```  

* After initial build, you can update the image with an adjusted system configuration script, without building Armbian from scratch:
  ```
  $ make update
  ```

* To clean up and remove the build environment, run `make clean`.

## Common issues

### Not enough disk space

The build needs 25G+. On Linux, the VirtualBox data directory is `~/VirtualBox VMs` by default,
so the host user running the build will need sufficient storage space in their home directory.

### Filesystem doesn't support permissions

If using a filesystem like exFAT, UNIX file permissions are not supported and the build will
not succeed. Move the repo and the virtualbox data directory to a filesystem that does support
file permissions to proceed.