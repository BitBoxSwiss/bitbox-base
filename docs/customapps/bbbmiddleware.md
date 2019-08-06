---
layout: default
title: Middleware
nav_order: 110
parent: Custom applications
---
## BitBox Base: Middleware

The secure communication channel between BitBox App and Base needs to carry a multitude of data exchanges: system health information, node management, Bitcoin blockchain queries to the Electrum server, Lightning client management and even more applications in the future. This communication must be flexibly routable under varying circumstances and protected with end-to-end encryption. By channeling all communication through a single endpoint on each side, clients do not need to support all native API protocols; authentication and encryption is handled only once.

On the BitBox Base, our custom Base Middleware (written in Go, open-source) is responsible to manage initial pairing, authentication, encryption and data distribution to native RPC and API interfaces. This allows for a more stable interface, additional functionality and potentially multiple user groups (like read-only access for friends & family).

One caveat of this approach, specifically tailored for ease-of-use, is its proprietary nature. This is where "advanced settings" will come in, allowing experienced users to open the native ports and gain full root access.

We have a `middleware` application written in Go, which exposes a single endpoint to the network, acting as a server when the user connects using the [BitBox App](https://github.com/digitalbitbox/bitbox-wallet-app/):

[`middleware/`](https://github.com/digitalbitbox/bitbox-base/tree/master/middleware)
