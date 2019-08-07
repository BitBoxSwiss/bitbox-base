---
layout: default
title: Security considerations
parent: Operating System
nav_order: 130
---
## Security considerations

As a networked device, reachable from any network, the attack surface of the BitBox Base needs to be minimized as much as possible. This is why we use our open-source [BitBox App](https://github.com/digitalbitbox/bitbox-wallet-app/) as the dedicated user interface for node management. By default, only one port is exposed on the Base to communicate over an end-to-end encrypted channel with the App, providing all functionality.

### Pairing Base with BitBox App

The first step is to safely pair the BitBox Base and App. To achieve this, the Base announces itself on the local network using mDNS and the Middlware exposes provides an API endpoint to initiate a secure connection. The Base is automatically detected by the App and announced within the user interface.

At first, only a single public API endpoint for pairing is available that allows the two components to establish an encrypted communication channel. To rule out a man-in-the-middle attack, a confirmation code is shown both within the App and on the Base and needs to be confirmed manually. With the secure connection established, the [Noise encryption](http://noiseprotocol.org) can be set up and the full API functionality, like node management, Bitcoin transaction verfication or Lightning Network usage, is available to the App.
  
### Networking

The following components help harden the Base against networking attacks:

* **No browser interface**: Modern web applications, built conveniently with frameworks like Node.js or Flask, have hundreds of dependencies that are outside our control and could contain security vulnerabilities or even malicious code. This is why we decided not to provide a web interface directly on the Base and expose only a single API by default.

* **NGINX**: although the Middleware exposes only a single API, additional components like electrs can optionally provide their own public API, e.g. for direct usage with the Electrum wallet. To have only a single entry-point from the public network into the Base, all communication is routed through [NGINX](https://www.nginx.com), which is a enterprise-grade, reliable and secure reverse proxy server.

* **Firewall**: to mitigate against network snooping and the exploitation of potentially open ports on operating system level, very restrictive packet filtering rules are set with [iptables](https://netfilter.org/projects/iptables/index.html) to refuse any connections that are not explicitely allowed.

* **Brute-force protection**: by default, SSH-login is disabled. Once enabled, only login with SSH keys should be used. Nonetheless, we use [fail2ban](https://www.fail2ban.org) to log any login attempts and block ip addresses for a certain time after too many unsuccessful tries.

* **SSH Keys**: for secure usage of SSH, the password login is disabled by default, so that a terminal login is only possible using SSH keys. The public key can be integrated in the Base image on build time, or later with the App.  
  
* **Disabled root user**: SSH login for the "root" user is disabled by default, even using SSH keys. This way, even when gaining access to the device using a stolen key, the attacker still needs to enter the "base" user password when using `sudo`.

* **Service users**: every service runs as a dedicated service user, with as few privileges as possible.

### Local security

A network appliance like the BitBox Base, built on readily available components, can never be assumed to be safe against an attacker that has physical access to the device. Nonetheless, the device should be safe against casual tampering.

* **Disabled users**: both `root` and `base` users are locked by default, and can only be unlocked using the BitBox App over the secure API after setting a new password.

* **Disabled USB**: to prevent tampering with a USB keyboard, the USB ports are configured to ignore USB-HID by default.

This list of security controls is non-exhaustive and meant to grow over time. If you think of additional controls, please let us know by opening a [GitHub issue](https://github.com/digitalbitbox/bitbox-base/issues). If you think you discovered a critical security issue, please contact us directly following the procedure of responisble disclosure outlined in our [bug bounty program](https://shiftcrypto.ch/bug-bounty-program/).
