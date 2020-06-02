---
layout: default
title: Method
parent: Updates
nav_order: 100
---
## Update method

In terms of upgrading a system, there are several approaches:

* **script-based, no modularization**: upgrading the system with scripts, without modularization of individual components, is a lost cause as differences and issues accumulate over time.

* **modular updates**: a more reliable approach is to encapsulate components in individual modules, e.g. using Docker.
  Every module can be updated to a defined state.
  Issues can arise when different versions of modules are not working well together, or when the base operating system needs updating.

* **disk image update**: updates overwrite the whole disk partition, usually containing the operating system and applications.
  This results in a very reliable, cleanly defined state.
  The whole system can be tested in-depth beforehand, but this low-level method adds complexity to the update process.

### Our approach: atomic disk image updates

We are working towards the goal to build the BitBoxBase as an appliance with 'firmware', not as a small Linux server.
This is why we decided to implement a full disk image update process, using the [Mender](https://mender.io/) open-source solution.

This solution has the following features:

* **enterprise-grade**: despite being open-source and free to use, Mender is a professional solution with its own Deployment Management server and a robust client update daemon.
* **full disk image**: the whole operating system including all applications are updated together, resulting in a clearly defined state of the system.
* **atomic**: the update process is atomic, as it either succeeds in full (including custom application tests), or not at all. In case of failure, a full rollback is performed.
* **on-demand**: the BitBoxApp will prompt you when an update is available, but you're not forced to update.
* **secure**: all updates are provided over a secure TLS communication connection and are cryptographically signed. By default, no unsigned updates are accepted by the device.
* **efficient**: update images are compressed and are streamed directly to the device. As of today, a full update (operating system and all applications) is ~300 MB in size.
* **minimal downtime**: the update is performed in the background, while applications continue to run. Only a single reboot is necessary to start the updated configuration.
* **persistent data**: the eMMC contains multiple partitions, one of which is used to store persistent data like the device configuration.
* **custom Base image**: more in the #reckless category, updating custom-built Base image can be enabled through the BitBoxApp.

Additional information is available at <https://mender.io> and their [technical documentation](https://docs.mender.io).
