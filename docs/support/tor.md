---
layout: default
title: Tor
parent: Supporting Applications
nav_order: 720
---
## Tor: Private Networking

Running the BitBox Base means connecting to other Bitcoin and Lightning Network peers.
If done over the regular internet protocol (IPv4), your IP address is visible to all your communication partners.
As IP addresses can be located geographically, this means revealing your physical location, from the general area down to individual city blocks (the granularity depends on how your internet service provider manages IP addresses).
You can easily locate yourself using services like <https://iplocation.io> or <http://www.infosniper.net>.

Revealing your location with the information that you are using Bitcoin is not a good idea.
This is why the BitBox Base uses [The Onion Router](https://www.torproject.org/) (Tor) to communicate to other Bitcoin peers privately and without revealing your real IP address.

By creating "hidden services", Tor can also be used to securely access your BitBox Base from outside your local network without having to configure any networking devices, such as port-forwarding on your router.

Please not that only Bitcoin-related communication is routed over Tor.

### Installation

Tor is installed using signed Debian packages from the official repository at `https://deb.torproject.org/torproject.org`.

### Configuration

The Tor configuration is application-specific:

* **Bitcoin Core**
  * Currently, `bitcoind` is not configured to listen to other Tor nodes.
  * To connect to other Bitcoin nodes, all communication is routed over the Tor SOCKS proxy. This configuration is set in `/etc/bitcoin/bitcoin.conf` with the option `proxy=127.0.0.1:9050`.
  Peers will see a regular IP address, but one that belongs to the Tor network and not your own.
  * For initial peer discovery on the Tor network, well-known `.onion` nodes are provided to get additional addresses from:  

    ```bash
    seednode=nkf5e6b7pl4jfd4a.onion
    seednode=xqzfakpeuvrobvpj.onion
    seednode=tsyvzsqwa2kkf6b2.onion
    ```

* **c-lightning**
  * `lightningd` is configured to use the Tor SOCKS proxy, as specified in `/etc/lightningd/lightningd.conf` with `proxy=127.0.0.1:9050`.
  * It can be reached from the outside using Tor hidden services, that are specified in `/etc/tor/torrc` (the `#LN#` comments allow scripts to identify these lines). Both a version 2 and 3 hidden service is created:  

    ```bash
    HiddenServiceDir /var/lib/tor/lightningd-service_v2/      #LN2#
    HiddenServicePort 9375 127.0.0.1:9735                     #LN2#

    HiddenServiceDir /var/lib/tor/lightningd-service_v3/      #LN#
    HiddenServiceVersion 3                                    #LN#
    HiddenServicePort 9375 127.0.0.1:9735                     #LN#
    ```

* **electrs**
  * Electrs does not include advanced networking features and relies on other applications for these, e.g. on NGINX to provide SSL encryption, or a correctly configured Tor hidden service.
  * Access to `electrs` is configured in `/etc/tor/torrc`:  

    ```bash
    HiddenServiceDir /var/lib/tor/hidden_service_electrum/    #ELECTRUM#
    HiddenServiceVersion 3                                    #ELECTRUM#
    HiddenServicePort 50002 127.0.0.1:50002                   #ELECTRUM#
    ```

* **Middleware**
  * The Base Middleware API is also accessible as a Tor hidden service, configured in `/etc/tor/torrc`:

    ```bash
    HiddenServiceDir /var/lib/tor/hidden_service_middleware/  #MIDDLEWARE#
    HiddenServiceVersion 3                                    #MIDDLEWARE#
    HiddenServicePort 9375 127.0.0.1:8845                     #MIDDLEWARE#
    ```

* **SSH**
  If enabled (off by default), SSH login is also available using a Tor hidden service, configured in `/etc/tor/torrc`:
  
  ```bash
  HiddenServiceDir /var/lib/tor/hidden_service_ssh/         #SSH#
  HiddenServiceVersion 3                                    #SSH#
  HiddenServicePort 22 127.0.0.1:22                         #SSH#
  ```

### Hidden services hostnames

Each hidden service is represented as a directory in `/var/lib/tor/` that contains files storing public and secret keys as well as a dedicated hostname (ending in `.onion`).

The individual hostnames are stored in the file `hostname`.

### Service management

Tor is started and managed by systemd, but no manual configuration of the unit is required, as the standard configuration activated at package installation is used.
