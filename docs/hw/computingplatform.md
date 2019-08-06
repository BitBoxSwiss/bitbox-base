---
layout: default
title: Computing platform
parent: Hardware 
nav_order: 100
---
## Computing platform

The BitBox Base aims to be a best-in-class solution, providing important functionalities for Bitcoin, while being able to handle additional use-cases related to digital and financial sovereignty in the future. The hardware therefore needs to be reliable and future-proof, with the possibility to upgrade individual parts like data storage.

The **RockPro64** is one of the most powerful ARM boards. Key features include:

* fast multi-core CPU (2 Cortex-A72 for performance + 4 Cortex A53 for low-power usage)
* 4 GB LDDR4 memory: enough memory for demanding programs that is also very fast, accelerating initial block download significantly
* eMMC support, providing more durable storage than microSD based solutions
* PCIe slot for internal storage connection
* great board layout with technical ports (ethernet, power, display) and user-facing ports (USB / power switch) at opposite ends
* real barrel power jack (not a flimsy USB port)
* designed as Long Term Supply (LTS) available until at least 2023

For speed, resilience and silent operations the BitBox Base uses an **SSD M.2 drive** that can be mounted on a PCIe adapter which connects directly to the PCIe slog on the board. As no adapter was available that minimizes the height once mounted, we produced our own SSD adapter to not waste any space.

An **integrated HSM**, an adapted version of the upcoming BitBox 02 hardware wallet, will drive a **trusted screen and capacitive touch buttons**. The PCB is optimized for mounting behind a glass cover. This component is especially interesting, as it allows for new use-cases like automatic signing for transaction that meet certain criteria, without exposing the private keys to the networked device itself. The display itself will also be able to display "untrusted" information from the Armbian firmware, but that needs to follow certain rules and be made clearly visible.

During normal operation, the board does not need to be cooled actively, running completely silent. During initial block download, however, the processor needs active cooling as it otherwise throttles down to avoid thermal damage. A **medium heatsink with integrated fan**, controlled by our Base Supervisor takes care that.

Additional components are a 16GB **eMMC module** (that can be flashed with a handy USB adapter by Pine64) and a **power supply**, which surprisingly can be quite low-powered (e.g. 12V/3A) due to the humble SSD drive. For the first low quantity batches, the BitBox Base **cases are 3D printed**. Due to the integrated HSM, which is attached to a glass cover, the final product will have a premium-feel nonetheless.

All in all this hardware platform provides enough power to complete the initial block download of the whole Bitcoin blockchain over Tor, including verification of all transactions since the Genesis block, in less than two days.

### Integrated HSM

A networked device can never be viewed as truly secure. Therefore, adding a Hardware Security Module (HSM) can drastically improve the security for use-cases that depend on trusted information or a need to safeguard secrets. We plan to integrate HSM functionality based on a modified version of the BitBox02 hardware wallet. This allows for safe pairing and can enable future use-cases like automated transaction signing to whitelisted addresses.

The HSM will run Bitcoin-only firmware and incorporates a trusted display and capacitive touch buttons. Together with the possibility to show limited "untrusted" data from the Linux system, with the trust-level clearly indicated, the screen and buttons provide the flexibility to create a great user experience. Additionally, LEDs on the HSM can allow for a quick way to show additional information like the overall status of the Base.
