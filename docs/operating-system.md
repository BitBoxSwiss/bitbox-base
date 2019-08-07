---
layout: default
title: Operating System
nav_order: 400
has_children: true
permalink: /os
---
# BitBox Base: Operating System

As the BitBox Base is designed as a networked appliance, it's important that the base operating system is reliable, but also heavily customizable. Specialized solutions to build customizable Linux environment like Buildroot or the Yocto project would be a good fit, but due to lack of hardware support and the immense complexity of these suites, we decided on a similar but much simpler approach. A later move to an "enterprise-grade" embedded Linux distribution is possible.

The Debian-based Linux distribution Armbian also features a reliable build environment that allows to build the operating system from source, configure the resulting kernel in minute detail and customize the resulting disk image using regular bash scripts within a chroot environment.

More detail on how to build the base operating image yourself will be detailed in the [Armbian build](armbian-build.md) docs.
