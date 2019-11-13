---
layout: default
title: Additional scripts
parent: Custom applications
nav_order: 160

---
## Additional scripts

Several helper scripts are located in the [`/opt/shift/scripts/`](https://github.com/digitalbitbox/bitbox-base/tree/master/armbian/base/scripts) directory.

### [**bbb-systemctl.sh**](https://github.com/digitalbitbox/bitbox-base/blob/master/armbian/base/scripts/bbb-systemctl.sh): manage and check systemd units in batch
Batch control all systemd units at once, e.g. for getting an overall status or stop all services.
```
BitBoxBase: batch control system units
Usage: bbb-systemctl.sh <status|start|restart|stop|enable|disable|verify>
```

(TODO)Stadicus
