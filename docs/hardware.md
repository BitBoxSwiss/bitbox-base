---
layout: default
title: Hardware
nav_order: 300
has_children: true
permalink: /hardware
---
# BitBox Base: Hardware

The BitBox Base aims to be a best-in-class solution, providing important functionalities for Bitcoin, while being able to handle additional use-cases related to digital and financial sovereignty in the future. The hardware therefore needs to be reliable and future-proof, with the possibility to upgrade individual parts like data storage.

![BitBox Base: Protoype photo](bbb-photo.jpg)  

The solution is comprised of the following main components:

* Computing: Pine64 RockPRO64 single board computer, hexa-core CPU, 4 GB DDR4 memory
* Storage: PCIe M.2 SSD drive for internal storage
* Security: BitBox secure module, with trusted OLED screen and capacitive buttons
* Case: custom enclosure with glass top, backmounted display

This preliminary schematic shows the combination of these components:

![BitBox Base: Schematic exploded](bbb-schematic.png)  

All in all this hardware platform provides enough power to download and verify the whole Bitcoin blockchain over Tor in less than two days. During the initial block download active cooling is essential, as both CPU and SSD are under heavy load. Once in regular operations mode, the BitBox Base is able to run very quietly.
