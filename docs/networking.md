---
layout: default
title: Networking
nav_order: 700
---
# BitBoxBase: Networking

Consistently working connectivity of devices without manual configuration is one of the main challenges of using a personal Bitcoin node. Poor privacy practices like revealing your own IP address (which can be geo-located quite accurately to your probable home area), the need to configure your router for external access and usage of unencrypted connections on public networks puts both your security and privacy at risk.

The BitBoxBase plans to solve this issue comprehensively, by providing three connection methods:

1. **Local network**
   The BitBoxBase announces its services using mDNS within the local network. When the BitBox App is connected to the same local network, be it at home or in an enterprise setting (without firewalls interfering), it automatically discovers the Base and can either initiate the setup process (with necessary physical pairing through the integrated secure module) or establish a trusted, encrypted connection when already paired.

2. **Tor network**
   Once initialized, the BitBoxBase automatically creates a Tor Hidden Service with an onion address. This service is announced to the Tor network and connections can then be established from the public internet. As the initial service connection is established from within the local network to the outside world and that connection is then kept alive, there is no need for any router configuration. Again, additional end-to-end encryption is used to secure all communication.

   One drawback of this method is that the computer running the BitBox App must have Tor installed, so that the BitBox App can route all traffic through the available Tor proxy. That might not be possible on every device.

3. **Shift Connect**
   For maximum flexibility, BitBoxBase users will be able to connect through the Shift Connect proxy that provides a regular HTTPS endpoint for each registered node, configured automatically on pairing. This service is purely opt-in and can not gather any information as the BitBoxBase connects with Tor to the Shift Connect proxy, without revealing its identity, ip address or location.

   As the proxy effectively is a man-in-the-middle, SSL encryption does not add enough privacy guarantees and is used only to secure the connection from the App to the proxy. This is why all data routed over the proxy is end-to-end encrypted. The public ip address connecting to the proxy is the only information that theoretically could be logged (which obviously won't be the case).

All communication channels are encrypted using the Noise Protocol Framework, from the BitBox App backend to the BitBoxBase middleware. On the Base, traffic is routed through the NGINX reverse-proxy and terminated at the custom Base Middleware.
