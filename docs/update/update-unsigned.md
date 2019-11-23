---
layout: default
title: Unsigned images
parent: Updates
nav_order: 800
---
## Update unsigned images

It is possible to update the BitBoxBase with unsigned update artefacts, but this is disabled by default for security reasons.

At the moment, it's only possible to enable this feature from the command line (e.g. over SSH):

```sh
sudo bbb-config.sh enable unsigned_updates
```

After that, an update artefact can be installed from a USB flashdrive.
The file must be located in the root folder and be named `update.base`.
To mount the flashdrive, use the following commands that are also used for backup purposes.

```sh
$ sudo bbb-cmd.sh flashdrive check
/dev/sda1

$ sudo bbb-cmd.sh flashdrive mount /dev/sda1
FLASHDRIVE MOUNT: mounted /dev/sda1 to /mnt/backup

$ sudo bbb-cmd.sh mender-update install flashdrive
```

The setting `unsigned_updates` is not persisted over updates on purpose, so if the new image was built with the setting turned off, it needs to be enabled again for a next update.
