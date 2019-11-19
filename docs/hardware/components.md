---
layout: default
title: Components
parent: Hardware
nav_order: 100
---
## Components

### Computing

The **RockPro64** is one of the most powerful ARM boards.
Key features include relevant to this use-case include:

* fast multi-core CPU (2 Cortex-A72 for performance + 4 Cortex A53 for low-power usage)
* 4 GB LDDR4 memory: enough memory for demanding programs that is also very fast, accelerating initial block download significantly
* eMMC support, providing more durable storage than microSD based solutions
* PCIe slot for internal storage connection
* great board layout with technical ports (ethernet, power, display) and user-facing ports (USB / power switch) at opposite ends
* real barrel power jack (not a flimsy USB port)
* designed as Long Term Supply (LTS) available until at least 2023

During normal operation, the board does not need to be cooled actively, running completely silent.
During initial block download, however, the processor needs active cooling as it otherwise throttles down to avoid thermal damage.
A **medium heatsink with integrated fan**, controlled by our Base Supervisor takes care that.

### Storage

For speed, resilience and silent operations the BitBoxBase uses an **SSD M.2 drive** that can be mounted on a PCIe adapter which connects directly to the PCIe slog on the board.
As no adapter was available that minimizes the height once mounted, we produced our own SSD adapter to not waste any space.
Standard adapter work as well.

Additional components are a 16GB **eMMC module** and a **power supply**, which surprisingly can be quite low-powered (e.g. 12V/3A) due to the humble SSD drive.

### Security

A networked device can never be viewed as truly secure.
Therefore, adding a secure module can drastically improve the security for use-cases that depend on trusted information or a need to safeguard secrets.
An **integrated BitBox secure module**, an adapted version of the [BitBox02 hardware wallet](https://shiftcrypto.ch/bitbox02/), will drive a **trusted screen and capacitive touch buttons**.
The PCB with the 2.4" OLED (128x64 px) screen is optimized for mounting behind glass.

This component, running a modified Bitcoin-only BitBox firmware, allows for new use-cases like automatic signing for transaction that meet certain criteria, without exposing the private keys to the networked device itself.
The display itself will also be able to display "untrusted" information from the Armbian Base image, but that needs to follow certain rules and be made clearly visible.
Additionally, LEDs on the secure module can allow for a quick way to show additional information like the overall status of the Base.

### Case

For the first low quantity batches, the BitBoxBase **cases are 3D printed** with a size of approximately 95 x 142 mm.
Due to the tinted glass cover, the final product will have a premium-feel nonetheless.
