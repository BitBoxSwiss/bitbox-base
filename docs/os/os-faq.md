---
layout: default
title: Common issues
parent: Operating system
nav_order: 900
---
## Common issues

### Not enough disk space

The build needs several gigabytes of disk space. On Linux, the Docker data directory is `/var/lib/docker` by default, so this directory needs to be on a filesystem with sufficient space.

### Armbian build fails

Building the Armbian within Docker is a bit tricky.
There are known issues creating loopback devices, which are used to create the disk images and partitions.

If the build fails at the very end like this, it is usually enough either reboot your computer and run `make build-update` again.

To clean up the loopback devices, you can run `contrib/cleanup-loop-devices.sh`(https://github.com/digitalbitbox/bitbox-base/blob/master/contrib/cleanup-loop-devices.sh), but please check the code first to avoid misconfiguring your host system.
To do this automatically, create an empty trigger file with `touch armbian/.cleanup-loop-devices`.
