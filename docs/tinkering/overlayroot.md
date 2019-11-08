---
layout: default
title: Overlayroot filesystem
parent: Tinkering
nav_order: 300
---
## Overlayroot filesystem

You must be aware that the whole **root filesystem is mounted as read-only**, and a temporary in-memory filesystem is used as an overlay. That means that you can change anything you want on the rootfs of the BitBoxBase (not the SSD, though!), but **on reboot, all changes are gone**.

This is great for resilience, eg. no corruption can occur on a sudden power outage, but it's not great for tinkering with your device.

For quick changes to the read-only filesystem, you can chroot into it:

```
$ sudo overlayroot-chroot
INFO: Chrooting into [/media/root-ro]
...
# do your stuff
...
$ exit
```

To deactivate the overlay root filesystem, run `sudo bbb-config.sh disable overlayrootfs` from the command line. After a reboot, the rootfs is mounted regularly and all changes are persistent. You can always `enable` the overlayrootfs again, using the same script. Each official update overwrites the whole rootfs, however.

If you build the BitBoxBase image yourself, you can configure the option `BASE_OVERLAYROOT` in [build.conf](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build.conf).
