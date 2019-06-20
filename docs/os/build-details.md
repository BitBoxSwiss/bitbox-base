---
layout: default
title: Build details
parent: Operating System
nav_order: 315
---
## Armbian build details

This document goes into more detail on how the Armbian build process works.

The [`armbian/build.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/build.sh) script is the entrypoint to the Armbian build process.
When running `make` in the `armbian/` directory, it in turn invokes the `armbian/build.sh` script under the hood.

It performs the following steps:

1. cloning the [`github.com/armbian/build`](https://github.com/armbian/build/) repo into `armbian/armbian-build/`, if it doesn't exist already
1. copying over the following files from the host system into `armbian/armbian-build/userpatches/overlay/`:
    1. `armbian/base/`: scripts and configs included in the Armbian image
    1. `build/`: contains all [built Go binaries](/go-apps.md) and their associated `.service` files
    1. [`customize-image.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/customize-image.sh): the hook which the Armbian build process calls
1. constructing the appropriate build arguments
1. calling the [`armbian/armbian-build/compile.sh`](https://github.com/armbian/build/blob/master/compile.sh) script with the `docker` argument, to kick off the dockerized build process

During the Armbian build process that's started by [`compile.sh`](https://github.com/armbian/build/blob/master/compile.sh), our [`customize-image.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/customize-image.sh) script is eventually called inside the build container.
This script copies over scripts and configs from the overlay directory into the Armbian filesystem, and then calls our [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/customize-armbian-rockpro64.sh) script.

Finally, the [`customize-armbian-rockpro64.sh`](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build/customize-armbian-rockpro64.sh) script handles steps like:

1. creating users/groups to run the different services
1. configuring SSH keys
1. installing runtime dependencies
1. creating and enabling systemd `.service` files
1. writing miscellaneous configs
