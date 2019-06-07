![BitBox Base logo](bitbox-base-logo.png)

[![Build Status](https://travis-ci.org/digitalbitbox/bitbox-base.svg?branch=master)](https://travis-ci.org/digitalbitbox/bitbox-base)

The BitBox Base is an ongoing project of [Shift Cryptosecurity](https://shiftcrypto.ch/) that aims to build a personal Bitcoin & Lightning full node appliance. The software is completely open-source and can be adapted to other hardware platforms.

All details outlined here reflect the current progress of the project and are subject to change if necessary.

## Technical documentation

The in-depth [technical documentation](https://digitalbitbox.github.io/bitbox-base) is hosted seperately and is updated frequently.

## Project goals

We believe that storing Bitcoin private keys on a hardware wallet like the [BitBox](https://shiftcrypto.ch) is only one part of the equation to gain financial sovereignty. While hardware wallets provide security, they do not provide privacy. Your entire financial history can be read by the company, such as the hardware wallet provider, who queries the blockchain for you. Because we respect an individual's right to privacy, we decided to build the BitBox Base. The currently missing part of the equation is a personal appliance that syncs directly with the Bitcoin peer-to-peer network and is able to send and validate transactions in a private manner. Trusting a third party to check your current Bitcoin balance is to be avoided.

The goals of BitBox Base:

* Running your own Bitcoin full node is for everyone.
* The built-in Lightning Network backend provides a compelling Lightning Wallet in the BitBox App.
* Connecting to your node just works, whether in your own network or on-the-go.
* Privacy is assured through comprehensive end-to-end encryption between User Interface and BitBox Base.
* As a networked appliance, remote attack surface is minimized by exposing as little ports as possible.
* The hardware platform uses best-in-class components, built for performance and resilience.
* With an integrated personal HSM, the node offers functionality previously not possible with hardware wallets.
* With atomic upgrades, upgrades of the firmware are seamless and reliable.
* Expert settings allow access to low-level configuration.

## Architecture

The BitBox Base integrates seamlessly with the BitBox App, which functions as control center for all node functionality, and supported hardware wallets. The two components discover each other within a local network without manual configuration and can then reconnect after initial pairing using different connection methods. This way, no webserver needs to be exposed and port-forwarding can be avoided completely.

The BitBox App is free and open-source, there's no need to buy a BitBox hardware wallet (although we would like that very much).

Building a solution platform that focuses on security and performance, the BitBox Base uses an ARM-based board with enough processing power to enable additional features in the future.

* [Pine64 ROCKPro64](https://www.pine64.org/rockpro64/) with fast 4GB memory and SSD
* [Armbian](https://www.armbian.com/): custom built Linux operating system, running on eMMC to facilitate atomic firmware updates
* Personal HSM: adapted BitBox 02 with trusted screen and buttons

The following base-level applications are used:

* [Bitcoin Core](https://bitcoincore.org/): full Bitcoin node, communicating directly with the peer-to-peer network, validating and broadcasting transactions
* [c-lightning](https://github.com/ElementsProject/lightning/blob/master/README.md): Lightning Network client specifically built for backend usage
* [electrs](https://github.com/romanz/electrs/blob/master/README.md): Electrum Server to provide blockchain data to software wallets

The following services are exposed:

* [NGINX](https://www.nginx.com/): reverse proxy to handle all incoming traffic
* Base Middleware: custom middleware to communicate with BitBox App, creating an encrypted communication channel and providing node management, Lightning client APIs as well as HSM integration
* Base Supervisor: custom daemon for operational monitoring and control, providing system health information and node configuration

Additional noteworthy components on the BitBox Base:

* [Tor](https://www.torproject.org/): external network connections exclusively use the privacy-focused Tor network
* [Prometheus](https://prometheus.io/): monitoring of system and software components
* [Grafana](https://grafana.com/): visualization of system and network performance metrics
* [Mender](https://mender.io/): atomic updates of filesystem images over-the-air

### Hardware considerations

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

An **integrated HSM**, an adapted version of the upcoming BitBox02 hardware wallet, will drive a **trusted screen and capacitive touch buttons**. The PCB is optimized for mounting behind a glass cover. This component is especially interesting, as it allows for new use-cases like automatic signing for transaction that meet certain criteria, without exposing the private keys to the networked device itself. The display itself will also be able to display "untrusted" information from the Armbian firmware, but that needs to follow certain rules and be made clearly visible.

During normal operation, the board does not need to be cooled actively, running completely silent. During initial block download, however, the processor needs active cooling as it otherwise throttles down to avoid thermal damage. A **medium heatsink with integrated fan**, controlled by our Base Supervisor takes care that.

Additional components are a 16GB **eMMC module** (that can be flashed with a handy USB adapter by Pine64) and a **power supply**, which surprisingly can be quite low-powered (e.g. 12V/3A) due to the humble SSD drive. For the first low quantity batches, the BitBox Base **cases are 3D printed**. Due to the integrated HSM, which is attached to a glass cover, the final product will have a premium-feel nonetheless.

All in all this hardware platform provides enough power to complete the initial block download of the whole Bitcoin blockchain over Tor, including verification of all transactions since the Genesis block, in less than two days.

### Operating System

As the BitBox Base is designed as a networked appliance, it's important that the base operating system is reliable, but also heavily customizable. Specialized solutions to build customizable Linux environment like Buildroot or the Yocto project would be a good fit, but due to lack of hardware support and the immense complexity of these suites, we decided on a similar but much simpler approach. A later move to an "enterprise-grade" embedded Linux distribution is possible.

The Debian-based Linux distribution Armbian also features a reliable build environment that allows to build the operating system from source, configure the resulting kernel in minute detail and customize the resulting disk image using regular bash scripts within a chroot environment.

More detail on how to build the base operating image yourself will be detailed in the [Armbian build](https://digitalbitbox.github.io/bitbox-base/os/armbian-build.html) docs.

### Integrated HSM

A networked device can never be viewed as truly secure. Therefore, adding a Hardware Security Module (HSM) can drastically improve the security for use-cases that depend on trusted information or a need to safeguard secrets. We plan to integrate HSM functionality based on a modified version of the BitBox02 hardware wallet. This allows for safe pairing and can enable future use-cases like automated transaction signing to whitelisted addresses. 

The HSM will run Bitcoin-only firmware and incorporates a trusted display and capacitive touch buttons. Together with the possibility to show limited "untrusted" data from the Linux system, with the trust-level clearly indicated, the screen and buttons provide the flexibility to create a great user experience. Additionally, LEDs on the HSM can allow for a quick way to show additional information like the overall status of the Base. 

### Networking

Consistently working connectivity of devices without manual configuration is one of the main challenges of using a personal Bitcoin node. Poor privacy practices like revealing your own IP address (which can be geo-located quite accurately to your probable home area), the need to configure your router for external access and usage of unencrypted connections on public networks puts both your security and privacy at risk.

The BitBox Base plans to solve this issue comprehensively, by providing three connection methods:

1. **Local network**  
   The BitBox Base announces its services using mDNS within the local network. When the BitBox App is connected to the same local network, be it at home or in an enterprise setting (without firewalls interfering), it automatically discovers the Base and can either initiate the setup process (with necessary physical pairing through the integrated HSM) or establish a trusted, encrypted connection when already paired.

2. **Tor network**  
   Once initialized, the BitBox Base automatically creates a Tor Hidden Service with an onion address. This service is announced to the Tor network and connections can then be established from the public internet. As the initial service connection is established from within the local network to the outside world and that connection is then kept alive, there is no need for any router configuration. Again, additional end-to-end encryption is used to secure all communication.

   One drawback of this method is that the computer running the BitBox App must have Tor installed, so that the BitBox App can route all traffic through the available Tor proxy. That might not be possible on every device.

3. **Shift Connect**  
   For maximum flexibility, BitBox Base users will be able to connect through the Shift Connect proxy that provides a regular HTTPS endpoint for each registered node, configured automatically on pairing. This service is purely opt-in and can not gather any information as the BitBox Base connects with Tor to the Shift Connect proxy, without revealing its identity, ip address or location. 

   As the proxy effectively is a man-in-the-middle, SSL encryption does not add enough privacy guarantees and is used only to secure the connection from the App to the proxy. This is why all data routed over the proxy is end-to-end encrypted. The public ip address connecting to the proxy is the only information that theoretically could be logged (which obviously won't be the case).

All communication channels are encrypted using the Noise Protocol Framework, from the BitBox App backend to the BitBox Base middleware. On the Base, traffic is routed through the NGINX reverse-proxy and terminated at the custom Base Middleware.

### Base Middleware

The secure communication channel between BitBox App and Base needs to carry a multitude of data exchanges: system health information, node management, Bitcoin blockchain queries to the Electrum server, Lightning client management and even more applications in the future. This communication must be flexibly routable under varying circumstances and protected with end-to-end encryption. By channeling all communication through a single endpoint on each side, clients do not need to support all native API protocols; authentication and encryption is handled only once.

On the BitBox Base, our custom Base Middleware (written in Go, open-source) is responsible to manage initial pairing, authentication, encryption and data distribution to native RPC and API interfaces. This allows for a more stable interface, additional functionality and potentially multiple user groups (like read-only access for friends & family).

One caveat of this approach, specifically tailored for ease-of-use, is its proprietary nature. This is where "advanced settings" will come in, allowing experienced users to open the native ports and gain full root access.

## Documentation

Detailed documentation is available at https://digitalbitbox.github.io/bitbox-base/.

## Contributor workflow

We are building the software stack of the BitBox Base fully open source and with its application outside of our own hardware device in mind. Contributions are very welcome. Please read [CONTRIBUTING](CONTRIBUTING.md) before submitting changes to the repository.


