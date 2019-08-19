---
layout: default
title: Supervisor
nav_order: 120
parent: Custom applications
---
## BitBox Base Supervisor

Every running system needs to be managed.
On a services level, this is the job of `systemd` which starts application services in the right order and restarts them on failure.
On top of this default service management, the BitBox Base Supervisor (`bbbsupervisor`) monitors the system using custom logic for the many intricacies of the various application components.
It follows application logs and checks system metrics in Prometheus, detects predefined events and when changes to the system become necessary, and is then able to trigger custom actions.

### Technical documentation

[See Docs on GitHub](https://github.com/digitalbitbox/bitbox-base/blob/master/tools/bbbsupervisor/README.md){: .btn }

### Scope

| System / Monitor | Action to take | Comment | OK |
|:--- |:--- |:--- | ---: |
| disk space full: check Prometheus metrics against threshold | Compact log files, user warning | Disk space ok for 5+ years | ☐ |
| ramdisk full: check Prometheus metric | restart device | active logrotate should prevent that | ☐ |
| swap file: periodically check `swapon -s` | recreate swap file | critical on new ssd setup, no issues expected after | ☐ |
| System temperature too high | none, part of `bbbfancontrol` |  | ✅ |
| **Middleware** ||||
| auth to bitcoind fails: `GetBlockChainInfo rpc call failed` | restart `bbbmiddleware.service` | bitcoin .cookie auth out of sync | ✅ |
| log `Failed to start c-lightning daemon.` and 'restart counter' over threshold. | user warning | OK during Bitcoin IBD |  ☐ |
| **Bitcoin** ||||
| change in "initial blockchain download" flag | update `bitcoin:ibd` in Redis | other services check if IBD is finished | ✅ |
| -"- | change `dbcache` to either `2000` (during IBD) or `300` MB | speed up IBD, free up memory otherwise | ☐ |
| log `No space left on device` | user warning, SSD full | Disk space ok for 5+ years | ☐ |
| log `Failed to start Bitcoin daemon.` and 'restart counter' over threshold. | user warning | OK during Bitcoin IBD |  ☐ |
| **Lightning** ||||
| channel backup | recurring (not possible yet) |  | ☐ |
| log `Failed to start c-lightning daemon.` and 'restart counter' over threshold. | user warning | OK during Bitcoin IBD | ☐ |
| **Electrs** ||||
| log `finished full compaction` | restart `electrs.service` | free memory after initial indexing | ☐ |
| auth to bitcoind fails: log `reconnecting to bitcoind: no reply from daemon` | restart `electrs.service` to update auth info | bitcoin .cookie auth out of sync | ☐ |
| log `Failed to start Electrs server daemon.` and 'restart counter' over threshold. | user warning | OK during Bitcoin IBD | ☐ |
