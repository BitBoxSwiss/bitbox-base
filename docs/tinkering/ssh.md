---
layout: default
title: SSH access
parent: Tinkering
nav_order: 200
---
## SSH access

SSH access is disabled by default.
Once enabled, you should always log in with the user `base` that has sudo privileges.
The user password corresponds to the one set in the BitBoxApp setup wizard.

There are multiple ways to gain access, some usable for production, others only suitable to be used for development:

* **SSH keys**: if SSH keys are present in `/home/base/.ssh/authorized_keys`, SSH login is possible over regular IP address, the mDNS domain (e.g. `ssh base@bitbox-base.local`) or even a Tor hidden service (if enabled).

  Currently, the keys need to be added manually, either by logging in locally or after login in with a password (see next option).
  We plan to allow users to add SSH keys from the BitBoxApp.

* **Password login**: this authentication method is not secure and should not be enabled for longer periods on a production device.

  It can be enabled in the BitBoxApp node management under "Advanced options".
  Alternatively, you can run `sudo bbb-config.sh enable sshpwlogin` directly on the command line.
  After enabling, you can log in with the user `base` using the password set in the Setup wizard.

* **Root login**: SSH access for the `root` user is disabled by default.
  For development, it can be enabled from the command line, e.g. to copy updated scripts directly into system folders that require root access.
  On the BitBoxBase, logged in with user `base`, run `sudo bbb-config.sh enable rootlogin`.

If you build the BitBoxBase image yourself, you can configure the options `BASE_LOGINPW` (initial login password, overwritten by the Setup Wizard) and `BASE_SSH_PASSWORD_LOGIN` in [build.conf](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/build.conf).
