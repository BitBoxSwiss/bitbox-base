---
layout: default
title: Factory reset
parent: Tinkering
nav_order: 400
---
## Factory reset

It is possible to reset various aspects the BitBoxBase to factory settings, either through the BitBoxApp (not implemented yet) or by creating a trigger file on the backup USB flashdrive.

In case of a forgotten password (set in the BitBoxApp Setup Wizard), it can be reset as follows:

* On the backup flashdrive, create a file named `reset-base-auth`.
* The flashdrive must contain a valid reset token, created on initial setup or subsequent backups
* Plug the flashdrive into the BitBoxBase
* Restart the device
* The authentication is now reset, run the BitBoxApp Setup Wizard again to set a new password.

Other reset options will be implemented in the future.
